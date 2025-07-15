package path

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"

	"github.com/muidea/magicCommon/foundation/log"
	fu "github.com/muidea/magicCommon/foundation/util"
)

type Op uint32

const (
	Create Op = 1 << iota
	Modify
	Remove
)

func (s Op) String() string {
	switch s {
	case Create:
		return "Create"
	case Modify:
		return "Modify"
	case Remove:
		return "Remove"
	}
	return "Unknown"
}

type Event struct {
	Path string
	Op   Op
}

type Observer interface {
	OnEvent(event Event)
}

// 监控指定目录及其子目录下的文件变化
// 包括新增、删除和修改
type Monitor struct {
	ignores      fu.StringSet
	syncMutex    sync.Mutex
	watchedPaths fu.StringSet
	fsWatcher    *fsnotify.Watcher
	observer     []Observer
}

func NewMonitor(ignores fu.StringSet) (*Monitor, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Monitor{
		ignores:      ignores,
		watchedPaths: fu.StringSet{},
		fsWatcher:    watcher,
	}, nil
}

func (s *Monitor) Start() error {
	eventQueue := make(chan fsnotify.Event, 1000)
	go func() {
		for event := range eventQueue {
			s.handleEvent(event)
		}
	}()
	go func() {
		for {
			select {
			case event, ok := <-s.fsWatcher.Events:
				if !ok {
					return
				}
				select {
				case eventQueue <- event:
				default:
					log.Warnf("event queue full, dropping event：%v", event)
				}
			case err, ok := <-s.fsWatcher.Errors:
				if !ok {
					return
				}
				_ = err
			}
		}
	}()

	return nil
}

func (s *Monitor) Stop() error {
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()

	s.observer = []Observer{}
	s.watchedPaths = fu.StringSet{}
	return s.fsWatcher.Close()
}

func (s *Monitor) AddObserver(observer Observer) {
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()

	s.observer = append(s.observer, observer)
}

func (s *Monitor) RemoveObserver(observer Observer) {
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()

	for i, o := range s.observer {
		if o == observer {
			s.observer = append(s.observer[:i], s.observer[i+1:]...)
			break
		}
	}
}

func (s *Monitor) AddIgnore(ignores fu.StringSet) {
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()

	for _, ignore := range ignores {
		s.ignores = s.ignores.Add(ignore)
	}
}

func (s *Monitor) isIgnore(path string) bool {
	for _, v := range s.ignores {
		if strings.Contains(path, v) {
			return true
		}
	}
	return false
}

func (s *Monitor) AddPath(path string) error {
	go s.refresh(path)
	return nil
}

func (s *Monitor) RemovePath(path string) error {
	return s.removePath(path)
}

func (s *Monitor) handleEvent(event fsnotify.Event) {
	if s.isIgnore(event.Name) {
		return
	}

	//log.Infof("op:%s, path:%s", event.Op, event.Name)

	if s.isDir(event.Name) {
		s.pathEvent(event)
		return
	}

	s.fileEvent(event)
}

func (s *Monitor) isDir(path string) bool {
	// 如果文件存在，则直接判断文件是否是目录，否则判断文件是否在目录缓存中
	pathInfo, pathErr := os.Stat(path)
	if pathErr == nil {
		return pathInfo.IsDir()
	}

	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()
	return s.watchedPaths.Exist(path)
}

func (s *Monitor) pathEvent(event fsnotify.Event) {
	switch event.Op {
	case fsnotify.Create:
		s.addPath(event.Name)
	case fsnotify.Write:
		return
	case fsnotify.Remove, fsnotify.Rename:
		s.removePath(event.Name)
		return
	default:
		return
	}
	filepath.WalkDir(event.Name, func(path string, info os.DirEntry, err error) error {
		if err != nil || s.isIgnore(path) {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		s.syncMutex.Lock()
		defer s.syncMutex.Unlock()
		for _, observer := range s.observer {
			observer.OnEvent(Event{
				Path: path,
				Op:   Create,
			})
		}
		return nil
	})
}

func (s *Monitor) fileEvent(event fsnotify.Event) {
	var localEvent Event
	switch event.Op {
	case fsnotify.Create:
		localEvent = Event{
			Path: event.Name,
			Op:   Create,
		}
		s.addPath(filepath.Dir(event.Name))
	case fsnotify.Write:
		localEvent = Event{
			Path: event.Name,
			Op:   Modify,
		}
	case fsnotify.Remove, fsnotify.Rename:
		localEvent = Event{
			Path: event.Name,
			Op:   Remove,
		}
	default:
		return
	}

	for _, observer := range s.observer {
		observer.OnEvent(localEvent)
	}
}

func (s *Monitor) addPath(path string) error {
	addFunc := func(addPath string) {
		if s.isIgnore(addPath) {
			return
		}
		s.syncMutex.Lock()
		defer s.syncMutex.Unlock()

		if s.watchedPaths.Exist(addPath) {
			return
		}

		//log.Warnf("add path:%s", addPath)
		if err := s.fsWatcher.Add(addPath); err != nil {
			return
		}
		s.watchedPaths = s.watchedPaths.Add(addPath)
	}

	addFunc(path)
	filepath.WalkDir(path, func(subPath string, info os.DirEntry, err error) error {
		if err != nil || s.isIgnore(subPath) {
			return nil
		}
		if info.IsDir() {
			addFunc(subPath)
		}

		addFunc(filepath.Dir(subPath))
		return nil
	})

	return nil
}

func (s *Monitor) removePath(path string) error {
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()

	var toDelete []string
	for _, subPath := range s.watchedPaths {
		if IsSubPath(path, subPath) {
			toDelete = append(toDelete, subPath)
		}
	}

	for _, subPath := range toDelete {
		//log.Warnf("remove path: %s", subPath)
		s.watchedPaths = s.watchedPaths.Remove(subPath)
		if err := s.fsWatcher.Remove(subPath); err != nil {
			return err
		}
	}

	return nil
}

func (s *Monitor) refresh(path string) {
	s.addPath(path)
	filepath.WalkDir(path, func(path string, info os.DirEntry, err error) error {
		if err != nil || s.isIgnore(path) {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		s.syncMutex.Lock()
		defer s.syncMutex.Unlock()
		for _, observer := range s.observer {
			observer.OnEvent(Event{
				Path: path,
				Op:   Create,
			})
		}
		return nil
	})
}

package path

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"

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
	watchedPaths map[string]bool
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
		watchedPaths: make(map[string]bool),
		fsWatcher:    watcher,
	}, nil
}

func (s *Monitor) Start() error {
	go func() {
		for {
			select {
			case event, ok := <-s.fsWatcher.Events:
				if !ok {
					return
				}
				s.handleEvent(event)
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
	s.watchedPaths = make(map[string]bool)
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

func (s *Monitor) AddPath(path string) error {
	go s.refresh(path)
	return nil
}

func (s *Monitor) RemovePath(path string) error {
	return s.removePath(path)
}

func (s *Monitor) handleEvent(event fsnotify.Event) {
	for _, v := range s.ignores {
		if strings.Contains(event.Name, v) {
			return
		}
	}

	var localEvent Event
	switch event.Op {
	case fsnotify.Create:
		localEvent = Event{
			Path: event.Name,
			Op:   Create,
		}
		if IsDir(event.Name) {
			s.addPath(event.Name)
		} else {
			s.addPath(filepath.Dir(event.Name))
		}
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
		if IsDir(event.Name) {
			s.removePath(event.Name)
		}
	default:
		return
	}

	if IsDir(event.Name) {
		return
	}

	for _, observer := range s.observer {
		observer.OnEvent(localEvent)
	}
}

func (s *Monitor) addPath(path string) error {
	addFunc := func(addPath string) {
		for _, v := range s.ignores {
			if strings.HasSuffix(addPath, v) {
				return
			}
		}

		s.syncMutex.Lock()
		defer s.syncMutex.Unlock()

		if s.watchedPaths[addPath] {
			return
		}

		if err := s.fsWatcher.Add(addPath); err != nil {
			return
		}
		s.watchedPaths[addPath] = true
	}

	addFunc(path)
	filepath.WalkDir(path, func(subPath string, info os.DirEntry, err error) error {
		if err != nil {
			return err
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
	for subPath := range s.watchedPaths {
		if strings.HasPrefix(subPath, path) {
			toDelete = append(toDelete, subPath)
		}
	}

	for _, subPath := range toDelete {
		if !s.watchedPaths[subPath] {
			return nil
		}

		if err := s.fsWatcher.Remove(subPath); err != nil {
			return err
		}
		delete(s.watchedPaths, subPath)
	}

	return nil
}

func (s *Monitor) refresh(path string) {
	s.addPath(path)
	filepath.WalkDir(path, func(path string, info os.DirEntry, err error) error {
		for _, v := range s.ignores {
			if strings.Contains(path, v) {
				return nil
			}
		}

		if err != nil {
			return err
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

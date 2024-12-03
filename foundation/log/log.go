package log

import (
	"encoding/xml"
	"os"
	"strconv"

	log "github.com/cihub/seelog"
)

const (
	levelAll      = -1
	levelTrace    = 0
	levelDebug    = 1
	levelInfo     = 2
	levelWarn     = 3
	levelError    = 4
	levelCritical = 5
	levelNone     = 6
)

func init() {
	var err error
	logger, err := log.LoggerFromConfigAsBytes([]byte(logConfig))
	if err != nil {
		logger.Criticalf("Failed to initialize logger: %v", err)
		os.Exit(1)
	}

	if err := validateLogConfig(logConfig); err != nil {
		logger.Criticalf("Invalid log configuration: %v", err)
		os.Exit(1)
	}

	defer func() {
		logger.SetAdditionalStackDepth(1)
		log.ReplaceLogger(logger)
	}()

	levelVal, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		logLevel = levelAll
		logger.Warn("LOG_LEVEL environment variable not set, defaulting to all levels")
		return
	}

	iVal, iErr := strconv.Atoi(levelVal)
	if iErr != nil {
		logLevel = levelTrace
		logger.Warnf("Invalid LOG_LEVEL value '%s', defaulting to trace level", levelVal)
		return
	}
	if iVal < levelTrace {
		logLevel = levelAll
		logger.Warnf("LOG_LEVEL value '%d' is too low, defaulting to all levels", iVal)
		return
	}
	if iVal > levelCritical {
		logLevel = levelNone
		logger.Warnf("LOG_LEVEL value '%d' is too high, defaulting to none level", iVal)
		return
	}

	logLevel = iVal
}

func validateLogConfig(config string) error {
	var xmlConfig struct{}
	if err := xml.Unmarshal([]byte(config), &xmlConfig); err != nil {
		return err
	}
	return nil
}

var logLevel int

var logConfig = `<?xml version="1.0" encoding="utf-8"?>
<seelog levels="trace,debug,info,warn,error,critical" type="sync"> 
  <outputs formatid="main"> 
    <!-- 对控制台输出的Log按级别分别用颜色显示。6种日志级别我仅分了三组颜色，如果想每个级别都用不同颜色则需要简单修改即可 -->  
    <filter levels="trace,debug,info"> 
      <console formatid="colored-default"/>  
    </filter>  
    <filter levels="warn"> 
      <console formatid="colored-warn"/>  
    </filter>  
    <filter levels="error,critical"> 
      <console formatid="colored-error"/>  
    </filter> 
  </outputs>  
  <formats> 
    <format id="colored-default" format="%EscM(38)%Date %Time [%LEV] %RelFile:%Line | %Msg%n%EscM(0)"/>  
    <format id="colored-warn" format="%EscM(33)%Date %Time [%LEV] %RelFile:%Line | %Msg%n%EscM(0)"/>  
    <format id="colored-error" format="%EscM(31)%Date %Time [%LEV] %RelFile:%Line | %Msg%n%EscM(0)"/>  
    <format id="main" format="%Date %Time [%LEV] %RelFile:%Line | %Msg%n"/> 
  </formats> 
</seelog>`

func Tracef(format string, params ...interface{}) {
	if logLevel > levelTrace {
		return
	}

	if len(params) > 0 {
		log.Tracef(format, params...)
		return
	}

	log.Trace(format)
}

func Debugf(format string, params ...interface{}) {
	if logLevel > levelDebug {
		return
	}

	if len(params) > 0 {
		log.Debugf(format, params...)
		return
	}

	log.Debug(format)
}

func Infof(format string, params ...interface{}) {
	if logLevel > levelInfo {
		return
	}

	if len(params) > 0 {
		log.Infof(format, params...)
		return
	}

	log.Info(format)
}

func Warnf(format string, params ...interface{}) {
	if logLevel > levelWarn {
		return
	}

	if len(params) > 0 {
		log.Warnf(format, params...)
		return
	}

	log.Warn(format)
}

func Errorf(format string, params ...interface{}) {
	if logLevel > levelError {
		return
	}

	if len(params) > 0 {
		log.Errorf(format, params...)
		return
	}

	log.Error(format)
}

func Criticalf(format string, params ...interface{}) {
	if logLevel > levelCritical {
		return
	}

	if len(params) > 0 {
		log.Criticalf(format, params...)
		return
	}

	log.Critical(format)
}

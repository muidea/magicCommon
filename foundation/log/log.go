package log

import (
	log "github.com/cihub/seelog"
)

func init() {
	logger, _ := log.LoggerFromConfigAsBytes([]byte(logConfig))
	logger.SetAdditionalStackDepth(1)
	log.ReplaceLogger(logger)
}

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
	if len(params) > 0 {
		log.Tracef(format, params...)
		return
	}

	log.Trace(format)
}

func Debugf(format string, params ...interface{}) {
	if len(params) > 0 {
		log.Debugf(format, params...)
		return
	}

	log.Debug(format)
}

func Infof(format string, params ...interface{}) {
	if len(params) > 0 {
		log.Infof(format, params...)
		return
	}

	log.Info(format)
}

func Warnf(format string, params ...interface{}) {
	if len(params) > 0 {
		log.Warnf(format, params...)
		return
	}

	log.Warn(params)
}

func Errorf(format string, params ...interface{}) {
	if len(params) > 0 {
		log.Errorf(format, params...)
		return
	}

	log.Error(format)
}

func Criticalf(format string, params ...interface{}) {
	if len(params) > 0 {
		log.Criticalf(format, params...)
		return
	}

	log.Critical(format)
}

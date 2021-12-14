package application

import (
	log "github.com/cihub/seelog"
)

func init() {
	logger, _ := log.LoggerFromConfigAsBytes([]byte(logConfig))
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

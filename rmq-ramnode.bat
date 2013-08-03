SETLOCAL
SET BASE=c:\temp\rmqramnode1
SET RMQHOME=C:\Program Files (x86)\RabbitMQ Server\rabbitmq_server-3.1.0
SET RABBITMQ_NODENAME=rabbit_ram1
SET RABBITMQ_BASE=%BASE%
SET RABBITMQ_NODE_PORT=5673
mkdir %BASE%
call "%RMQHOME%\sbin\rabbitmq-server.bat" -detached
call "%RMQHOME%\sbin\rabbitmq-plugins.bat" enable rabbitmq_management_agent
call "%RMQHOME%\sbin\rabbitmqctl.bat" stop_app
REM call "%RMQHOME%\sbin\rabbitmqctl.bat" reset
call "%RMQHOME%\sbin\rabbitmqctl.bat" join_cluster --ram rabbit@FAWAD-THINKPAD
call "%RMQHOME%\sbin\rabbitmqctl.bat" start_app
:END
ENDLOCAL
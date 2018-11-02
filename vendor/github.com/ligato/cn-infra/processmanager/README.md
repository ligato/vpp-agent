# Process manager

Process manager plugin provides a set of methods to start/stop process and read information about it. 

`NewProcess()` starts a process. Every process has name, command and a set of arguments. Created process
is then stored in the plugin memory and can be read via `GetProcess(<name>)`. To see all processes handled
by plugin, call `GetAll()`.

Created process is not started automatically. The process 'p' defines `p.Start()` method which executes given
command and sets internal status to **running**. The method does not directly return process ID. The value
is stored in object and can be obtained with `p.GetPID`. Other significant API methods:

* `Stop()` stops the process
* `Kill()` force-stops the process
* `Restart()` tries to stop the process gently and then stats it again. If process cannot be stopped, 
error is returned.
* `Wait()` waits till the process is completed and returns its status

Other methods can be used to check whether the process is alive, get its name, original command, argument,
time when it was started etc. 
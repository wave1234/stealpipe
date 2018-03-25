# stealpipe
最牛逼的安全传输软件

# Steal Pipe ！！
##	Steal Pipe 是一款开源的安全软件，它可以保护你的数据传输，让它无法被黑客查看和监控.
##	            同时它可以帮助你把本地端口和远端服务器端口连接在一起.

# 注意：

  1 ./pipe -h 可以获得参数的使用说明
  2 源码中的quic 经过了简单修改，数据进行了加密，不要用正常的版本替代它。
   

# 可以实现以下功能

 1. 2个服务器之间进行数据传输， 可以使用CBC HTTP QUIC 3种方式进行加密，其中CBC和HTTP利用TCP, QUIC 使用UDP,使用QUIC的优点是不会被RST影响。
 
 2. 在服务器上开启隐秘的端口，比如socks5 http服务。 比如 服务器上开启 http 服务，侦听127.0.0.1:80, 开启socks5 服务，侦听127.0.0.1:1080 端口，同时启动PIPE 服务，侦听 7777号UDP端口。客户端可以使用PIPE使用服务器的socks5服务，同时也能访问http服务器。服务器的http服务，如果没有PIPE的密码是无法访问的。同时也不会被外部的扫描工具扫描到，是非常安全的。
  
 
# 用法1： 本地和远方服务器进行数据传输
  本地ip： 192.168.1.1 本地使用8888端口. 远端服务器ip：10.0.0.1，远端服务器使用800-1000端口接受数据. 远端运行了web服务器 服务端口127.0.0.1:80 
  
  本地启动： pipe --key st1234 -r 1000 -s 300 --remotehost  10.0.0.1:800-1000  -port 8888 --pipetype client --encrypttype QUIC
  
  远端服务器启动： pipe --key st1234 -r 1000 -s 300 --remotehost  -port 800-1000 --remotehost 127.0.0.1:80 --encrypttype QUIC

# 用法2： 本地和远方服务器穿越本地HTTP防火墙进行数据传输
  本地ip： 192.168.1.1 本地使用8888， 远端服务器ip：10.0.0.1 远端服务器使用1000端口接受数据. 远端运行了web服务器 服务端口 127.0.0.1:80 
  
  本地启动进程1： pipe --remotehost  127.0.0.1:1234  -port 8888 --pipetype client
  
  本地启动进程2： pipe --remotehost  10.0.0.1:1000 --host 127.0.0.1 -port 1234 --pipetype client --encrypttype HTTP 

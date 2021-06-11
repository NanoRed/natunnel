# natunnel
This is an NAT tunnel implementation

## how it works
![arch](https://user-images.githubusercontent.com/23076538/121658456-f018cf00-cad3-11eb-915d-b7a2bedbbb07.png)
![flow](https://user-images.githubusercontent.com/23076538/121658017-8698c080-cad3-11eb-99eb-4ba9b341f581.png)

## hot to use
* server
```
$ git clone https://github.com/RedAFD/natunnel.git
$ cd natunnel
$ make server-linux
$ ./server -h
Usage of ./server:
  -HTTPParserAddr string
        Input the listening address of the http parser service. (default "127.0.0.1:7714")
  -HostDomain string
        Input Your Domain.
  -ServerAddr string
        Input the listening address of the natunnel server. (default "0.0.0.0:80")
$ nohup ./server -ServerAddr="0.0.0.0:80" -HostDomain="yourdomain.com" > /var/log/natunnel.log 2>&1 &
```
* client
```
PS C:\Users\XXX> git clone https://github.com/RedAFD/natunnel.git
PS C:\Users\XXX> cd natunnel
PS C:\Users\XXX> make client-windows // if your pc is macOS, use make client-darwin
PS C:\Users\XXX> .\client.exe
Please enter your natunnel server address(e.g. 40.100.70.1:80):
yourdomain.com:80
Please enter your local server address that need to be exposed to the internet(e.g. 127.0.0.1:8080):
127.0.0.1:80
Successfully running, your public host is http://7bb327.yourdomain.com. Enjoy :)
```

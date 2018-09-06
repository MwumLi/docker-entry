# docker-entry

使用 Docker Remote API 提供 Docker 容器远程终端服务

> 管理 Docker Container 不在本项目的范围之内, 这是 [Kubernetes](https://kubernetes.io/) 这种容器管理平台做的事情  
> 本项目旨在为容器 Web 云服务平台提供一个与已运行的 Docker Container 建立终端通信的后端服务  

## 使用的 Docker Remote API

本项目用到了 3 个 Docker Remote API:  

* [/containers/{id}/exec](https://docs.docker.com/engine/api/v1.33/#operation/ContainerExec): Create an exec instance, 返回一个 Exec instance ID  
  `id`: ID or name of container  

* [/exec/{id}/start](https://docs.docker.com/engine/api/v1.33/#operation/ExecStart): 启动之前创建的 exec instance, 得到执行结果或与之前建立流通信  
  * `id`:	Exec instance ID
  * **注意**: 因为我们要为 Docker 提供一个远程终端, 因此要使用 socket 通信的方式, 构造一个 post 请求, 与 Node 上的容器服务建立一个 Tcp 连接, 从而建立流通信  
  * [Hijacking](https://docs.docker.com/engine/api/v1.26/#operation/ContainerAttach)  

* [/exec/{id}/resize](https://docs.docker.com/engine/api/v1.33/#operation/ExecResize): Resize the TTY session used by an exec instance  
  * `id`:	Exec instance ID  
  * **Query Param**: `w` - 宽度 `h` - 高度, 单位都是字符  

## 搭配 xterm.js 使用

可以使用 [xterm.js](https://github.com/xtermjs/xterm.js/)  作为 Web 前端, 使用本项目作为后端, 实现一个 Web 容器远程终端, 从而在浏览器上操作向容器发送命令  

## 相关资源

* Docker Remote API 版本历史: <https://docs.docker.com/engine/api/version-history/>  
  此页可以看到 Docker Remote API 版本变更, 并提供各个 API 版本文档的入口  

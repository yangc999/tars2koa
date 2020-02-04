# API GATEWAY GUIDE

## 原理



## 部署

### Docker

1. 下载镜像

   * `docker pull postgres:9.6` 数据库版本限定,否则初始化会报错

   * `docker pull kong` api网关组件

   * `docker pull pantsel/konga` api网关监控

   * `docker pull springcloud/eureka` 服务注册中心

2. 虚拟网络

   `docker network create kong-net`

3. 部署postgres数据库

   `docker run -d --name kong-database --network=kong-net -p 5432:5432 -e "POSTGRES_USER=kong" -e "POSTGRES_DB=kong" postgres:9.6`

4. kong数据库初始化

   `docker run --rm --network=kong-net -e "KONG_DATABASE=postgres" -e "KONG_PG_HOST=kong-database" kong kong migrations bootstrap`
   
5. 部署kong

   `docker run -d --name kong --network=kong-net -e "KONG_DATABASE=postgres" -e "KONG_PG_HOST=kong-database" -e "KONG_ADMIN_LISTEN=0.0.0.0:8001, 0.0.0.0:8444 ssl" -p 8000:8000 -p 8443:8443 -p 8001:8001 -p 8444:8444 kong`

6. konga数据库初始化

   `docker run --rm --network=kong-net pantsel/konga -c prepare -a postgres -u postgresql://kong:@{{PostgresDockerIp}}:5432/konga`

7. 部署konga

   `docker run -p 1337:1337 --network kong-net --name konga -e "NODE_ENV=production" -e "DB_ADAPTER=postgres" -e "DB_URI=postgresql://kong:@{{PostgresDockerIp}}:5432/konga" pantsel/konga`

8. 部署eureka注册中心

   `docker run -d -p 8761:8761 springcloud/eureka`

9. 安装kong sync-eureka插件

   * `docker exec -it kong /bin/bash`

   * `luarocks install kong-plugin-sync-eureka`

   * `cp /etc/kong/kong.conf.default /etc/kong/kong.conf`

   * `vi /etc/kong/kong.conf`

   * 找到`plugins`选项,修改为`plugins = bundled,sync-eureka `,保存退出

   * `kong prepare & kong reload`

   * 在宿主机执行`curl -H "Content-Type: application/json" -X POST  --data '{"config":{"sync_interval":10,"eureka_url":"http://{{EurekaDockerIp}}:8761/eureka","clean_target_interval":86400},"name":"sync-eureka"}' http://{{KongDockerIp}}:8001/plugins`

     `curl -H "Content-Type: application/json" -X GET http://{{KongDockerIp}}:8001/plugins/`

## 使用

    ### Node.js

1. `npm install eureka-js-client --save`
2. 



   

   

   

   


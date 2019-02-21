#删除旧镜像
#docker rmi quay.io/coreos/etcd:v3.3.12
#下载镜像
docker pull quay.io/coreos/etcd:v3.3.12
#启动etcd服务
#--mount type=bind,source=$(pwd)\etcd_data\,destination=/etcd-data `
  docker run `
  -p 2379:2379 `
  -p 2380:2380 `
  --rm `
  --name etcd-gcr-v3.3.12 `
  quay.io/coreos/etcd:v3.3.12 `
  /usr/local/bin/etcd `
      --name s1 `
      --data-dir /etcd-data `
      --listen-client-urls http://0.0.0.0:2379 `
      --advertise-client-urls http://0.0.0.0:2379 `
      --listen-peer-urls http://0.0.0.0:2380 `
      --initial-advertise-peer-urls http://0.0.0.0:2380 `
      --initial-cluster s1=http://0.0.0.0:2380 `
      --initial-cluster-token tkn `
      --initial-cluster-state new



#执行客户端命令
docker exec etcd-gcr-v3.3.12 /bin/sh -c "/usr/local/bin/etcd --version"
docker exec etcd-gcr-v3.3.12 /bin/sh -c "ETCDCTL_API=3 /usr/local/bin/etcdctl version"
docker exec etcd-gcr-v3.3.12 /bin/sh -c "ETCDCTL_API=3 /usr/local/bin/etcdctl endpoint health"
docker exec etcd-gcr-v3.3.12 /bin/sh -c "ETCDCTL_API=3 /usr/local/bin/etcdctl put foo bar"
docker exec etcd-gcr-v3.3.12 /bin/sh -c "ETCDCTL_API=3 /usr/local/bin/etcdctl get foo"


#启动mongodb
docker run -p 27017:27017 --rm mongo:3.4 
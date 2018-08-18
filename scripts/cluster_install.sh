#! /bin/bash

# Generate the secret for the cluster

NODE="initial_peer"

# if we are setting up the first node, lets generate a new cluster secret
if [[ "$NODE" == "initial_peer" ]]; then
    CLUSTER_SECRET=$(od  -vN 32 -An -tx1 /dev/urandom | tr -d ' \n')
    export CLUSTER_SECRET=$(od  -vN 32 -An -tx1 /dev/urandom | tr -d ' \n')
fi

IPFS_CLUSTER_PATH=/ipfs/ipfs-cluster
export IPFS_CLUSTER_PATH=/ipfs/ipfs-cluster

# Install and initialize ipfs-cluster

cd ~ || exit
wget https://dist.ipfs.io/ipfs-cluster-service/v0.4.0/ipfs-cluster-service_v0.4.0_linux-amd64.tar.gz
tar zxvf ipfs-cluster-service*.tar.gz
rm -- *gz
cd ipfs-cluster-service || exit
./ipfs-cluster-service init
sudo cp ipfs-cluster-service /usr/local/bin
cd ~ || exit
wget https://dist.ipfs.io/ipfs-cluster-ctl/v0.4.0/ipfs-cluster-ctl_v0.4.0_linux-amd64.tar.gz
tar zxvf ipfs-cluster-ctl*.tar.gz
rm -- *gz
cd ipfs-cluster-ctl || exit
sudo cp ipfs-cluster-ctl /usr/local/bin
Sample code to provide backend endpoint for cloud-init phone-home functionality.

Build code:

```
wget https://golang.org/dl/go1.20.5.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.20.5.linux-amd64.tar.gz
echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.profile
source ~/.profile

mkdir cloudinit-phone-home
cd cloudinit-phone-home
go mod init cloudinit-phone-home

## Create maing.go with the contents from this repo and build
go build

## Sample output:

2024/05/26 22:26:41 Deleted instance with ID: 60001
2024/05/26 22:31:28 Created instance: {60001 create    success k0smotron-node-3   }
2024/05/26 22:31:34 Queried instance status: {60001 create    success k0smotron-node-3   }
2024/05/26 22:33:22 Created instance: {60001 create    success k0smotron-node-3   }
2024/05/26 22:33:34 Deleted instance with ID: 60001
2024/05/26 22:36:45 Created instance: {60001 create    success k0smotron-node-3   }
2024/05/26 22:36:50 Queried instance status: {60001 create    success k0smotron-node-3   }

You can check the sample TF code that utilizes this concept

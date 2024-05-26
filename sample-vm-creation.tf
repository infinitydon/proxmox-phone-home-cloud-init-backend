terraform {
  required_providers {
    proxmox = {
      source  = "bpg/proxmox"
      version = ">= 0.46.1"
    }
  }
}

resource "proxmox_virtual_environment_file" "cloud_config" {
  content_type = "snippets"
  datastore_id = var.snippets_storage_pool
  node_name    = var.proxmox_host_node

  source_raw {
    data = <<EOF
#cloud-config
disable_root: false
user: ubuntu
ssh_authorized_keys:
  - ${var.ssh_public_key}
chpasswd:
  expire: False
users:
  - default
  - name: ubuntu
    groups: sudo
    shell: /bin/bash
    sudo: ALL=(ALL) NOPASSWD:ALL
package_upgrade: true
package_reboot_if_required: true
packages:
- httpie
write_files:
  - path:  /etc/netplan/ens19-cloud-init.yaml
    permissions: '0644'
    content: |
         network:
           version: 2
           renderer: networkd
           ethernets:
             ens19:
               dhcp4: false
runcmd:
  - netplan generate
  - netplan apply
  - sed -i 's/[#]*PermitRootLogin prohibit-password/PermitRootLogin without-password/g' /etc/ssh/sshd_config
  - systemctl reload ssh
  - hostnamectl set-hostname `dmesg | grep -i k0s | awk '{print $5}' | sed 's/,//g'`
  - curl https://releases.rancher.com/install-docker/23.0.sh | sh
  - http --ignore-stdin POST http://192.168.100.62:8080/phone-home id=${var.vm_identifier} event_name=create name=${var.vm_name}
EOF

    file_name = "${var.vm_identifier}.cloud-config.yaml"
  }
}

resource "proxmox_virtual_environment_vm" "ubuntu_cloudinit_vms" {
  vm_id = var.vm_identifier

  name        = var.vm_name
  description = "k0s VMS"

  tags        = var.tags

  node_name = var.proxmox_host_node

  smbios {
    product = var.vm_name
  }  

  agent {
    enabled = true
  }
  on_boot = false

  clone {
    retries = 3
    vm_id   = var.os_template_id
  }

  disk {
    datastore_id = var.boot_disk_storage_pool
    interface    = "scsi0"
    file_format  = "raw"
    discard      = "on"
    size         = var.node_disk_size
  }

  initialization {
    ip_config {
      ipv4 {
        address = "dhcp"
      }
    }

    user_data_file_id = proxmox_virtual_environment_file.cloud_config.id
  }

  cpu {
    architecture = "x86_64"
    cores        = var.node_cpu_cores
    type         = "host"
  }
  memory {
    dedicated = var.node_memory
  }
  network_device {
    bridge  = var.config_network_bridge
    mac_address = var.nic_net0_mac_address
  }

  network_device {
    bridge  = var.config_network_bridge
  }  

  operating_system {
    type = "l26"
  }

  serial_device {} 

  lifecycle {
    ignore_changes = [
      disk[0].discard,
      vga
    ]
  }  
}

resource "null_resource" "wait_for_phone_home" {
  depends_on = [proxmox_virtual_environment_vm.ubuntu_cloudinit_vms]

  provisioner "local-exec" {
    command = <<EOT
      for i in {1..30}; do
        response=$(curl -s -w "HTTPSTATUS:%%{http_code}" "${var.check_status_url}?id=${var.vm_identifier}")
        http_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
        body=$(echo "$response" | sed -e 's/HTTPSTATUS:.*//')

        if [ "$http_code" -eq 200 ]; then
          status=$(echo "$body" | jq -r '.result')
          if [ "$status" == "success" ]; then
            echo "VM ${var.vm_identifier} cloud init has been completed"
            exit 0
          fi
        elif [ "$http_code" -eq 404 ]; then
          echo "Instance not found"
        else
          echo "Unexpected response: $body"
        fi

        sleep 10
      done
      exit 1
    EOT
    interpreter = ["/bin/bash", "-c"]
  }
}

resource "null_resource" "instance" {
  triggers = {
    instance_id = var.vm_identifier
    phone_home_url = var.phone_home_url
  }

  provisioner "local-exec" {
    when    = destroy
    command = "curl -X POST -H 'Content-Type: application/json' -d '{\"id\":\"${self.triggers.instance_id}\", \"event_name\": \"delete\"}' \"${self.triggers.phone_home_url}\""
  }
}

output "ip_addresses" {
  value     = proxmox_virtual_environment_vm.ubuntu_cloudinit_vms.ipv4_addresses
}

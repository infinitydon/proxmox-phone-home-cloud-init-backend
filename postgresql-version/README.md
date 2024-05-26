## Install Postgresql

apt-get install postgresql postgresql-contrib

## Get Golang postresql library, run this inside the same directory where main.go is located

go get github.com/lib/pq

## Loging to postresql and configure

```
-- Create database
CREATE DATABASE phone_home;

-- Grant all privileges on the database to the user
GRANT ALL PRIVILEGES ON DATABASE phone_home TO go_proxmox;

-- Connect to the database
\c phone_home

-- Create the instances table
CREATE TABLE instances (
    id SERIAL PRIMARY KEY,
    instance_id VARCHAR(255) NOT NULL,
    event_name VARCHAR(50),
    name VARCHAR(255),
    result VARCHAR(50)
);

-- Grant all privileges on all tables in the public schema to the user
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO go_proxmox;

-- Grant all privileges on the sequence associated with the id column
GRANT ALL PRIVILEGES ON SEQUENCE instances_id_seq TO go_proxmox;


\q
```

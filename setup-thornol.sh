#!/bin/sh

api_key="61cad96c4acfa972f334bd359c45b211d9cf4a6ee6e247e5ceb70eb7ec99e867"

# if opkg is not installed, abort
if ! command -v opkg &> /dev/null
then
  echo "opkg is not installed. Please install opkg and try again."
  exit 1
fi

# if base64 is not installed, install it
if ! command -v base64 &> /dev/null
then
  opkg update
  opkg install coreutils-base64
fi

# if jsonfilter is not installed, install it
if ! command -v jsonfilter &> /dev/null
then
  opkg update
  opkg install jsonfilter
fi

# if curl is not installed, install it
if ! command -v curl &> /dev/null
then
  opkg update
  opkg install curl
fi

. /usr/share/libubox/jshn.sh

download_thornol_binary(){
  # Download the binary using curl
  curl -o thornol.bin "https://binaries.scogo.golain.io/thornol_app"

  # Make the binary executable
  chmod +x thornol.bin

  # create directory if not exists
  mkdir -p /usr/lib/thornol

  # Move the binary to the directory
  mv thornol.bin /usr/bin/thornol
}

device_name=$(uci show scogo.@device[0].devicename | awk -F"'" '{print $2}')

project_id='995e9cfb-6bea-41e8-9f9c-ae110f367524'
org_id='a6c8f5fb-0f3c-443d-b616-a457501c9ed5'
fleet_id='607dbf4a-842b-4cad-96c7-ee07cf5e41ee'

create_new_device() {
  # Ensure api_key and device_name are defined
  if [ -z "$api_key" ] || [ -z "$device_name" ]; then
    echo "api_key or device_name is not set."
    exit 1
  fi
  # Register a new device based on a device template
  curl --location "https://dev.api.golain.io/core/api/v1/projects/$project_id/fleets/$fleet_id/devices/bulk" \
  --header "ORG-ID: $org_id" \
  --header "Content-Type: application/json" \
  --header "Authorization: APIKEY $api_key" \
  --data '{
    "device_count": 1,
    "fleet_device_template_id": "dc3be894-482d-4ffe-b9ca-ade2668e7ecf",
    "device_names": ["'"$device_name"'"]
  }' > response1.json

  # Check if the response is successful based on json file
  status=$(jsonfilter -i response1.json -e @.ok)
  if [ "$status" != "1" ]; then
    echo "Failed to register the device. Reason: $(jsonfilter -i response1.json -e @.message)"
    exit 1
  fi
}

provision_new_certificate_for_device(){
  # Extract the device id from the response
  device_id=$(jsonfilter -i response1.json -e @.data.deviceIds[0])

  # Ensure api_key and device_name are defined
  if [ -z "$api_key" ] || [ -z "$device_name" ] || [ -z "$device_id" ]; then
    echo "api_key or device_name is not set."
    exit 1
  fi
  # Provision new certificates for the device and decode the response from base64
  curl --location 'https://dev.api.golain.io/core/api/v1/projects/'"$project_id"'/fleets/'"$fleet_id"'/devices/'"$device_id"'/certificates' \
  --header 'ORG-ID: '"$org_id"'' \
  --header 'Content-Type: application/json' \
  --header 'Authorization: APIKEY '"$api_key"'' \
  --data '{}' > response2.json
}

extract_connection_settings(){
    # Load JSON data into jshn
  json_load_file response2.json

  # Navigate to the certificates object
  json_select certificates

  # Extract and decode connection settings
  json_get_var connection_settings "connection_settings.json"
  echo "$connection_settings" | base64 -d > connection_settings.json

  # Extract and decode device certificate
  json_get_var device_cert "device_cert.pem"
  echo "$device_cert" | base64 -d > device_cert.pem

  # Extract and decode device private key
  json_get_var device_private_key "device_private_key.pem"
  echo "$device_private_key" | base64 -d > device_private_key.pem

  # Extract and decode root CA certificate
  json_get_var root_ca_cert "root_ca_cert.pem"
  echo "$root_ca_cert" | base64 -d > root_ca_cert.pem

  mkdir -p /usr/lib/thornol/certs

  # move the certificates to the directory
  mv device_cert.pem /usr/lib/thornol/certs
  mv device_private_key.pem /usr/lib/thornol/certs
  mv root_ca_cert.pem /usr/lib/thornol/certs
  
  mv connection_settings.json /usr/lib/thornol
}

create_initd_service() {
  # Create a init.d service for the binary
  cat <<EOF > /etc/init.d/thornol
#!/bin/sh /etc/rc.common

START=99
STOP=10

USE_PROCD=1

start_service() {
    procd_open_instance thornol
    procd_set_param command /usr/bin/thornol

    procd_set_param limits core="unlimited"
    procd_set_param env GO_ENV=dev CONFIG_DIR=/usr/lib/thornol/
    procd_set_param stdout 1
    procd_set_param stderr 1
    procd_set_param respawn

    procd_set_param pidfile /var/run/thornol.pid
    procd_set_param user root
    procd_close_instance
}

restart_service() {
  stop
  start
}

EOF

  chmod +x /etc/init.d/thornol
  /etc/init.d/thornol enable

  /etc/init.d/thornol start
}

failure=1

# if the response1.json file does not exist or does not have the `ok` key set to 1
if [ ! -f response1.json ] || [ "$(jsonfilter -i response1.json -e @.ok)" != "1" ]; then
# and if the connection_settings file does not exist
  if [ ! -f /usr/lib/thornol/connection_settings.json ]; then
    create_new_device
  fi
fi

# if the response2.json file does not exist or does not have the `ok` key set to 1
if [ ! -f response2.json ] || [ "$(jsonfilter -i response2.json -e @.ok)" != "1" ]; then
# and if the device_private_key file does not exist
  if [ ! -f /usr/lib/thornol/certs/device_private_key.pem ]; then
    provision_new_certificate_for_device
  fi
fi

# extract the connection settings & certificates
if [ ! -f /usr/lib/thornol/certs/device_private_key.pem ]; then
  extract_connection_settings
fi

# replace the thornol binary if exists
if [ -f /usr/bin/thornol ]; then
  rm /usr/bin/thornol
fi

download_thornol_binary

# if the init.d service does not exist, create it
if [ ! -f /etc/init.d/thornol ]; then
  create_initd_service
  failure=0
else 
  /etc/init.d/thornol restart
  failure=0
fi


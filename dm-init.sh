alias dm=docker-machine

read "XO_URL?XO_URL: "
export XO_URL
read "XO_USERNAME?XO_USERNAME: "
export XO_USERNAME
read -s "XO_PASSWORD?XO_PASSWORD: "
export XO_PASSWORD
export XO_INSECURE=true
read "XO_TEMPLATE?XO_TEMPLATE: "
export XO_TEMPLATE
read "XO_VM_NETWORK?XO_VM_NETWORK: "
export XO_VM_NETWORK

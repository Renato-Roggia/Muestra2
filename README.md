# Muestra2
///////////////////////////////////////
# Actualizar sistema
sudo apt update
sudo apt upgrade -y

# Instalar dependencias permanentes
sudo apt install -y golang-go rabbitmq-server protobuf-compiler

# Configurar RabbitMQ para que inicie automáticamente
sudo systemctl enable rabbitmq-server
sudo systemctl start rabbitmq-server
////////////////////////////
nano ~/.bashrc
//////////////////
# Configuración permanente para Go
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
export GOBIN=$GOPATH/bin
/////////////////////////////
# Presiona Ctrl+X, luego Y, luego Enter para guardar
source ~/.bashrc
/////////////////////////////7
# Instalar los generadores de código
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Verificar instalación
which protoc-gen-go
which protoc-gen-go-grpc
//////////////////////////


cd /Users/admin/Учёба/AUnix/lb1/web_server/deploy
docker-compose up -d
cd ..
sleep 2 # wait for DB initialization
go run cmd/main.go
# Ctrl-C to quit

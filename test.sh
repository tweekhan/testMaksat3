
SERVER_URL="http://localhost:8088/ping"

CLIENT_IDS=("client1" "client2" "client3")


function send_ping {
  for ID in ${CLIENT_IDS[@]}; do
    curl "${SERVER_URL}/${ID}"
    echo " Пинг отправлен для $ID"
  done
}


while true; do
  send_ping
  sleep 65
done

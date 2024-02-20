import time
import json
import socket

MESSAGE: str = "Алиякбяров Марат Ансарович M30-310Б 21"
BUFFER_SIZE: int = 256

client: socket.socket = socket.socket(family=socket.AF_INET, type=socket.SOCK_STREAM)

with open(file="config.json", mode="r", encoding="utf-8") as config_file:
    config = json.load(config_file)
    address: str = config["address"]
    port: int = int(config["port"])
    log_file_name: str = config["log_file_name"]
    error_log_file_name: str = config["error_log_file_name"]
    timeout: int = config["timeout"]

log_file = open(file=log_file_name, mode="+a", encoding="utf-8")
error_log_file = open(file=error_log_file_name, mode="+a", encoding="utf-8")

try: 
    client.connect((address, port))
except Exception as error:
    error_log_file.write(f"{time.ctime()} - ошибка при подключении к серверу - {error}\n")
    raise ConnectionError (f"ошибка при подключении к серверу - {error}")

log_file.write(f"{time.ctime()} - успешное подключение - {address}:{port}\n")

time.sleep(timeout)

client.send(MESSAGE.encode())
log_file.write(f"{time.ctime()} - отправлено сообщение - {MESSAGE}\n")

time.sleep(timeout)

response_message: str = client.recv(BUFFER_SIZE).decode("utf-8")
log_file.write(f"{time.ctime()} - получен ответ от сервера - {response_message}\n")

client.close()
log_file.close()
error_log_file.close()
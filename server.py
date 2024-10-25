# server.py
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import redis
import time

app = FastAPI()
r = redis.Redis(host='localhost', port=6379, db=0)

TIMEOUT_SECONDS = 60

class Message(BaseModel):
    client_id: int
    content: str

@app.post("/send")
def send_message(message: Message):
    key = f"message:{message.client_id}"
    # Сохраняем сообщение с TTL из глобальной переменной
    r.setex(key, TIMEOUT_SECONDS, message.content)
    return {"status": "Message sent"}

@app.get("/receive/{client_id}")
def receive_message(client_id: int):
    key = f"message:{client_id}"
    start_time = time.time()
    while True:
        elapsed = time.time() - start_time
        if elapsed > TIMEOUT_SECONDS:
            return {"status": "No new messages"}
        message = r.get(key)
        if message:
            r.delete(key)
            return {"message": message.decode()}
        time.sleep(1)  # Ждем 1 секунду перед повторной проверкой

# Добавьте этот блок в конец файла
if __name__ == "__main__":
    import uvicorn
    uvicorn.run("server:app", host="0.0.0.0", port=8000)
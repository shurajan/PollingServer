import aiohttp
import asyncio
import time

CLIENT_ID = 2
OTHER_CLIENT_ID = 1
SERVER_URL = 'http://localhost:8000'

async def send_pong(session):
    async with session.post(f"{SERVER_URL}/send", json={
        "client_id": OTHER_CLIENT_ID,
        "content": "pong"
    }) as response:
        print(f"Client {CLIENT_ID} sent 'pong' to client {OTHER_CLIENT_ID}")

async def receive_messages():
    async with aiohttp.ClientSession() as session:
        while True:
            print(f"Client {CLIENT_ID} is waiting for messages...")
            try:
                async with session.get(f"{SERVER_URL}/receive/{CLIENT_ID}") as response:
                    data = await response.json()
                    if 'message' in data:
                        message = data['message']
                        print(f"Client {CLIENT_ID} received: {message}")
                        if message == 'ping':
                            await asyncio.sleep(1)
                            await send_pong(session)
                    else:
                        print(f"Client {CLIENT_ID}: No new messages.")
            except Exception as e:
                print(f"Client {CLIENT_ID}: Error occurred - {e}")
            await asyncio.sleep(0.1)

if __name__ == "__main__":
    asyncio.run(receive_messages())
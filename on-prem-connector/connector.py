"""
On-Prem Connector - Main Service
Connects to Railway backend via WebSocket and handles AD/Exchange requests
"""

import asyncio
import json
import logging
import os
import signal
import sys
from datetime import datetime
from typing import Optional

# Load .env file first
from dotenv import load_dotenv
load_dotenv()

import yaml
import websockets
from websockets.exceptions import ConnectionClosed

from ad_client import ADClient
from ews_client import EWSClient

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.StreamHandler(sys.stdout),
        logging.FileHandler('connector.log')
    ]
)
logger = logging.getLogger('connector')


class OnPremConnector:
    def __init__(self, config_path: str = 'config.yaml'):
        self.config = self._load_config(config_path)
        self.ad_client = ADClient(self.config)
        self.ews_client = EWSClient(self.config)
        self.ws: Optional[websockets.WebSocketClientProtocol] = None
        self.running = False
        self.reconnect_interval = self.config['backend'].get('reconnect_interval', 5)
        self.heartbeat_interval = self.config['backend'].get('heartbeat_interval', 30)

    def _load_config(self, path: str) -> dict:
        """Load and expand config with environment variables"""
        with open(path, 'r') as f:
            config = yaml.safe_load(f)

        # Expand environment variables
        def expand_env(obj):
            if isinstance(obj, str) and obj.startswith('${') and obj.endswith('}'):
                env_var = obj[2:-1]
                return os.environ.get(env_var, obj)
            elif isinstance(obj, dict):
                return {k: expand_env(v) for k, v in obj.items()}
            elif isinstance(obj, list):
                return [expand_env(item) for item in obj]
            return obj

        return expand_env(config)

    async def connect(self):
        """Connect to Railway backend via WebSocket"""
        url = self.config['backend']['url']
        api_key = self.config['backend']['api_key']

        # Add API key as query parameter (more compatible)
        if '?' in url:
            url = f"{url}&token={api_key}"
        else:
            url = f"{url}?token={api_key}"

        try:
            self.ws = await websockets.connect(
                url,
                ping_interval=20,
                ping_timeout=10
            )
            logger.info(f"Connected to backend: {url.split('?')[0]}")
            return True
        except Exception as e:
            logger.error(f"Failed to connect to backend: {e}")
            return False

    async def disconnect(self):
        """Disconnect from backend"""
        if self.ws:
            await self.ws.close()
            self.ws = None
            logger.info("Disconnected from backend")

    async def send_message(self, message: dict):
        """Send message to backend"""
        if self.ws:
            await self.ws.send(json.dumps(message))

    async def handle_message(self, message: str):
        """Handle incoming message from backend"""
        try:
            data = json.loads(message)
            command = data.get('command')
            request_id = data.get('request_id')
            params = data.get('params', {})

            logger.info(f"Received command: {command} (request_id: {request_id})")

            result = None
            error = None

            try:
                if command == 'ping':
                    result = {'pong': True, 'timestamp': datetime.now().isoformat()}

                elif command == 'sync_users':
                    result = self._handle_sync_users()

                elif command == 'get_user':
                    result = self._handle_get_user(params)

                elif command == 'authenticate':
                    result = self._handle_authenticate(params)

                elif command == 'get_subordinates':
                    result = self._handle_get_subordinates(params)

                elif command == 'get_calendar':
                    result = await self._handle_get_calendar(params)

                elif command == 'create_meeting':
                    result = await self._handle_create_meeting(params)

                elif command == 'update_meeting':
                    result = await self._handle_update_meeting(params)

                elif command == 'delete_meeting':
                    result = await self._handle_delete_meeting(params)

                elif command == 'find_free_slots':
                    result = await self._handle_find_free_slots(params)

                elif command == 'get_free_busy':
                    result = await self._handle_get_free_busy(params)

                else:
                    error = f"Unknown command: {command}"

            except Exception as e:
                logger.exception(f"Error handling command {command}")
                error = str(e)

            # Send response
            response = {
                'type': 'response',
                'request_id': request_id,
                'command': command,
                'success': error is None,
                'result': result,
                'error': error,
                'timestamp': datetime.now().isoformat()
            }
            await self.send_message(response)

        except json.JSONDecodeError as e:
            logger.error(f"Invalid JSON message: {e}")
        except Exception as e:
            logger.exception(f"Error processing message: {e}")

    def _handle_sync_users(self) -> dict:
        """Sync all users from AD"""
        users = self.ad_client.get_all_users()
        org_tree = self.ad_client.build_org_tree(users)
        return {
            'users': users,
            'org_tree': org_tree
        }

    def _handle_get_user(self, params: dict) -> Optional[dict]:
        """Get single user"""
        if 'email' in params:
            return self.ad_client.get_user_by_email(params['email'])
        elif 'dn' in params:
            return self.ad_client.get_user_by_dn(params['dn'])
        return None

    def _handle_authenticate(self, params: dict) -> dict:
        """Authenticate user against AD"""
        username = params.get('username')
        password = params.get('password')

        if not username or not password:
            return {'authenticated': False, 'error': 'Missing credentials'}

        user = self.ad_client.authenticate_user(username, password)
        if user:
            return {'authenticated': True, 'user': user}
        return {'authenticated': False, 'error': 'Invalid credentials'}

    def _handle_get_subordinates(self, params: dict) -> list:
        """Get subordinates for manager"""
        manager_dn = params.get('manager_dn')
        if not manager_dn:
            return []
        return self.ad_client.get_subordinates(manager_dn)

    async def _handle_get_calendar(self, params: dict) -> list:
        """Get calendar events"""
        email = params.get('email')
        if not email:
            return []
        return self.ews_client.get_calendar_events(
            email,
            days_back=params.get('days_back'),
            days_forward=params.get('days_forward')
        )

    async def _handle_create_meeting(self, params: dict) -> Optional[dict]:
        """Create meeting"""
        return self.ews_client.create_meeting(
            organizer_email=params.get('organizer_email'),
            subject=params.get('subject'),
            start=datetime.fromisoformat(params.get('start')),
            end=datetime.fromisoformat(params.get('end')),
            attendees=params.get('attendees', []),
            body=params.get('body', ''),
            location=params.get('location', '')
        )

    async def _handle_update_meeting(self, params: dict) -> Optional[dict]:
        """Update meeting"""
        return self.ews_client.update_meeting(
            organizer_email=params.get('organizer_email'),
            item_id=params.get('item_id'),
            subject=params.get('subject'),
            start=datetime.fromisoformat(params['start']) if params.get('start') else None,
            end=datetime.fromisoformat(params['end']) if params.get('end') else None,
            attendees=params.get('attendees'),
            body=params.get('body')
        )

    async def _handle_delete_meeting(self, params: dict) -> bool:
        """Delete meeting"""
        return self.ews_client.delete_meeting(
            organizer_email=params.get('organizer_email'),
            item_id=params.get('item_id')
        )

    async def _handle_find_free_slots(self, params: dict) -> list:
        """Find free time slots"""
        return self.ews_client.find_free_slots(
            emails=params.get('emails', []),
            duration_minutes=params.get('duration_minutes', 60),
            start=datetime.fromisoformat(params.get('start')),
            end=datetime.fromisoformat(params.get('end'))
        )

    async def _handle_get_free_busy(self, params: dict) -> dict:
        """Get free/busy info"""
        return self.ews_client.get_free_busy(
            emails=params.get('emails', []),
            start=datetime.fromisoformat(params.get('start')),
            end=datetime.fromisoformat(params.get('end'))
        )

    async def heartbeat_loop(self):
        """Send periodic heartbeat"""
        while self.running:
            try:
                if self.ws:
                    await self.send_message({
                        'type': 'heartbeat',
                        'timestamp': datetime.now().isoformat(),
                        'status': 'online'
                    })
                await asyncio.sleep(self.heartbeat_interval)
            except Exception as e:
                logger.error(f"Heartbeat error: {e}")
                break

    async def message_loop(self):
        """Main message receiving loop"""
        while self.running:
            try:
                if not self.ws:
                    break
                message = await self.ws.recv()
                await self.handle_message(message)
            except ConnectionClosed:
                logger.warning("WebSocket connection closed")
                break
            except Exception as e:
                logger.error(f"Message loop error: {e}")
                break

    async def run(self):
        """Main run loop with auto-reconnect"""
        self.running = True

        # Setup signal handlers
        loop = asyncio.get_event_loop()
        for sig in (signal.SIGINT, signal.SIGTERM):
            loop.add_signal_handler(sig, self.stop)

        logger.info("Starting On-Prem Connector...")

        # Initialize clients
        logger.info("Connecting to AD...")
        if self.ad_client.connect():
            logger.info("AD connection successful")
        else:
            logger.warning("AD connection failed - will retry on demand")

        logger.info("Connecting to Exchange...")
        if self.ews_client.connect_service_account():
            logger.info("Exchange connection successful")
        else:
            logger.warning("Exchange connection failed - will retry on demand")

        while self.running:
            try:
                if await self.connect():
                    # Run heartbeat and message loops concurrently
                    await asyncio.gather(
                        self.heartbeat_loop(),
                        self.message_loop()
                    )

                if self.running:
                    logger.info(f"Reconnecting in {self.reconnect_interval} seconds...")
                    await asyncio.sleep(self.reconnect_interval)

            except Exception as e:
                logger.error(f"Run loop error: {e}")
                if self.running:
                    await asyncio.sleep(self.reconnect_interval)

        # Cleanup
        await self.disconnect()
        self.ad_client.disconnect()
        logger.info("Connector stopped")

    def stop(self):
        """Stop the connector"""
        logger.info("Stopping connector...")
        self.running = False


def main():
    """Entry point"""
    config_path = os.environ.get('CONNECTOR_CONFIG', 'config.yaml')
    connector = OnPremConnector(config_path)

    try:
        asyncio.run(connector.run())
    except KeyboardInterrupt:
        pass


if __name__ == '__main__':
    main()

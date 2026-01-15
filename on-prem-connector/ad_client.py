"""
Active Directory Client (LDAP)
Handles user sync and org structure from AD
"""

import logging
from typing import Optional
from ldap3 import Server, Connection, ALL, NTLM, SUBTREE
from ldap3.core.exceptions import LDAPException
import base64

logger = logging.getLogger(__name__)


class ADClient:
    def __init__(self, config: dict):
        self.config = config['ad']
        self.conn: Optional[Connection] = None

    def connect(self) -> bool:
        """Establish connection to AD"""
        try:
            server = Server(
                self.config['server'],
                port=self.config['port'],
                use_ssl=self.config['use_ssl'],
                get_info=ALL
            )

            self.conn = Connection(
                server,
                user=self.config['bind_user'],
                password=self.config['bind_password'],
                authentication=NTLM,
                auto_bind=True
            )

            logger.info(f"Connected to AD: {self.config['server']}")
            return True

        except LDAPException as e:
            logger.error(f"Failed to connect to AD: {e}")
            return False

    def disconnect(self):
        """Close AD connection"""
        if self.conn:
            self.conn.unbind()
            self.conn = None
            logger.info("Disconnected from AD")

    def get_all_users(self, offset: int = 0, limit: int = 100, include_photo: bool = True) -> tuple[list[dict], int]:
        """Fetch users from AD with pagination"""
        if not self.conn:
            if not self.connect():
                return [], 0

        try:
            search_filter = "(&(objectClass=user)(objectCategory=person)(!(userAccountControl:1.2.840.113556.1.4.803:=2)))"

            self.conn.search(
                search_base=self.config['users_ou'],
                search_filter=search_filter,
                search_scope=SUBTREE,
                attributes=self.config['attributes']
            )

            all_entries = self.conn.entries
            total = len(all_entries)

            # Apply pagination
            paginated = all_entries[offset:offset + limit]

            users = []
            for entry in paginated:
                user = self._parse_user(entry, include_photo=include_photo)
                if user:
                    users.append(user)

            logger.info(f"Fetched {len(users)} users from AD (offset={offset}, limit={limit}, total={total})")
            return users, total

        except LDAPException as e:
            logger.error(f"Failed to fetch users: {e}")
            return [], 0

    def get_user_by_email(self, email: str) -> Optional[dict]:
        """Fetch single user by email"""
        if not self.conn:
            if not self.connect():
                return None

        try:
            search_filter = f"(&(objectClass=user)(mail={email}))"

            self.conn.search(
                search_base=self.config['base_dn'],
                search_filter=search_filter,
                search_scope=SUBTREE,
                attributes=self.config['attributes']
            )

            if self.conn.entries:
                return self._parse_user(self.conn.entries[0])
            return None

        except LDAPException as e:
            logger.error(f"Failed to fetch user {email}: {e}")
            return None

    def get_user_by_dn(self, dn: str) -> Optional[dict]:
        """Fetch user by Distinguished Name"""
        if not self.conn:
            if not self.connect():
                return None

        try:
            self.conn.search(
                search_base=dn,
                search_filter="(objectClass=user)",
                search_scope=SUBTREE,
                attributes=self.config['attributes']
            )

            if self.conn.entries:
                return self._parse_user(self.conn.entries[0])
            return None

        except LDAPException as e:
            logger.error(f"Failed to fetch user by DN {dn}: {e}")
            return None

    def get_subordinates(self, manager_dn: str) -> list[dict]:
        """Get direct reports for a manager"""
        if not self.conn:
            if not self.connect():
                return []

        try:
            search_filter = f"(&(objectClass=user)(manager={manager_dn}))"

            self.conn.search(
                search_base=self.config['base_dn'],
                search_filter=search_filter,
                search_scope=SUBTREE,
                attributes=self.config['attributes']
            )

            return [self._parse_user(entry) for entry in self.conn.entries if self._parse_user(entry)]

        except LDAPException as e:
            logger.error(f"Failed to fetch subordinates: {e}")
            return []

    def authenticate_user(self, username: str, password: str) -> Optional[dict]:
        """Authenticate user against AD"""
        try:
            server = Server(
                self.config['server'],
                port=self.config['port'],
                use_ssl=self.config['use_ssl']
            )

            # Try to bind with user credentials
            user_conn = Connection(
                server,
                user=username,
                password=password,
                authentication=NTLM,
                auto_bind=True
            )

            # If successful, get user info
            # Extract username without domain (can't use backslash in f-string)
            sam_account = username.split('\\')[-1] if '\\' in username else username
            user_conn.search(
                search_base=self.config['base_dn'],
                search_filter=f"(sAMAccountName={sam_account})",
                attributes=self.config['attributes']
            )

            if user_conn.entries:
                user = self._parse_user(user_conn.entries[0])
                user_conn.unbind()
                logger.info(f"User authenticated: {username}")
                return user

            user_conn.unbind()
            return None

        except LDAPException as e:
            logger.warning(f"Authentication failed for {username}: {e}")
            return None

    def _parse_user(self, entry, include_photo: bool = False) -> Optional[dict]:
        """Parse LDAP entry to user dict"""
        try:
            # Get photo as base64 only if requested (photos are large!)
            photo = None
            if include_photo and hasattr(entry, 'thumbnailPhoto') and entry.thumbnailPhoto.value:
                photo = base64.b64encode(entry.thumbnailPhoto.value).decode('utf-8')

            # Extract manager email from DN
            manager_dn = str(entry.manager) if hasattr(entry, 'manager') and entry.manager else None

            return {
                'dn': str(entry.entry_dn),
                'name': str(entry.cn) if hasattr(entry, 'cn') else None,
                'email': str(entry.mail) if hasattr(entry, 'mail') else None,
                'login': str(entry.sAMAccountName) if hasattr(entry, 'sAMAccountName') else None,
                'upn': str(entry.userPrincipalName) if hasattr(entry, 'userPrincipalName') else None,
                'title': str(entry.title) if hasattr(entry, 'title') else None,
                'department': str(entry.department) if hasattr(entry, 'department') else None,
                'manager_dn': manager_dn,
                'photo_base64': photo
            }
        except Exception as e:
            logger.error(f"Failed to parse user entry: {e}")
            return None

    def build_org_tree(self, users: list[dict]) -> dict:
        """Build organization tree from flat user list"""
        # Create lookup by DN
        by_dn = {u['dn']: u for u in users}

        # Add children lists
        for user in users:
            user['subordinates'] = []

        # Build tree
        roots = []
        for user in users:
            if user['manager_dn'] and user['manager_dn'] in by_dn:
                by_dn[user['manager_dn']]['subordinates'].append(user)
            else:
                roots.append(user)

        return {
            'roots': roots,
            'total_count': len(users)
        }


# Test connection
if __name__ == "__main__":
    import yaml
    import os

    logging.basicConfig(level=logging.DEBUG)

    with open('config.yaml', 'r') as f:
        config = yaml.safe_load(f)

    # Expand env vars
    config['ad']['bind_user'] = os.environ.get('AD_BIND_USER', config['ad']['bind_user'])
    config['ad']['bind_password'] = os.environ.get('AD_BIND_PASSWORD', config['ad']['bind_password'])

    client = ADClient(config)
    if client.connect():
        users = client.get_all_users()
        print(f"Found {len(users)} users")
        for u in users[:5]:
            print(f"  - {u['name']} ({u['email']})")
        client.disconnect()

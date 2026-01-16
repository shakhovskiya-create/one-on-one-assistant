"""
Exchange Web Services Client
Handles calendar sync and meeting management
"""

import logging
from typing import Optional
from datetime import datetime, timedelta
from exchangelib import (
    Credentials, Account, Configuration, DELEGATE,
    CalendarItem, Attendee, EWSDateTime, EWSTimeZone,
    Q
)
from exchangelib.protocol import BaseProtocol, NoVerifyHTTPAdapter

logger = logging.getLogger(__name__)

# Disable SSL verification for self-signed certs (common in on-prem)
# Remove this in production if you have proper certs
BaseProtocol.HTTP_ADAPTER_CLS = NoVerifyHTTPAdapter


class EWSClient:
    def __init__(self, config: dict):
        self.config = config['ews']
        self.sync_config = config.get('sync', {})
        self.accounts: dict[str, Account] = {}
        self.service_account: Optional[Account] = None

    def connect_service_account(self) -> bool:
        """Connect with service account for impersonation"""
        try:
            credentials = Credentials(
                username=self.config['username'],
                password=self.config['password']
            )

            config = Configuration(
                service_endpoint=self.config['url'],
                credentials=credentials,
                auth_type=self.config.get('auth_type', 'NTLM')
            )

            # Create account (will use impersonation for other users)
            self.service_account = Account(
                primary_smtp_address=self._extract_email(self.config['username']),
                config=config,
                autodiscover=False,
                access_type=DELEGATE
            )

            logger.info(f"Connected to Exchange: {self.config['url']}")
            return True

        except Exception as e:
            logger.error(f"Failed to connect to Exchange: {e}")
            return False

    def get_account(self, email: str) -> Optional[Account]:
        """Get or create account for user (with impersonation)"""
        if email in self.accounts:
            return self.accounts[email]

        try:
            credentials = Credentials(
                username=self.config['username'],
                password=self.config['password']
            )

            config = Configuration(
                service_endpoint=self.config['url'],
                credentials=credentials,
                auth_type=self.config.get('auth_type', 'NTLM')
            )

            account = Account(
                primary_smtp_address=email,
                config=config,
                autodiscover=False,
                access_type=DELEGATE
            )

            self.accounts[email] = account
            return account

        except Exception as e:
            logger.error(f"Failed to get account for {email}: {e}")
            return None

    def get_calendar_events(self, email: str, days_back: int = None, days_forward: int = None) -> list[dict]:
        """Get calendar events for a user"""
        account = self.get_account(email)
        if not account:
            return []

        days_back = days_back or self.sync_config.get('calendar_days_back', 7)
        days_forward = days_forward or self.sync_config.get('calendar_days_forward', 30)

        try:
            tz = EWSTimeZone.localzone()
            start = tz.localize(EWSDateTime.now() - timedelta(days=days_back))
            end = tz.localize(EWSDateTime.now() + timedelta(days=days_forward))

            events = []
            for item in account.calendar.filter(start__lt=end, end__gt=start).order_by('start'):
                events.append(self._parse_calendar_item(item))

            logger.info(f"Fetched {len(events)} calendar events for {email}")
            return events

        except Exception as e:
            logger.error(f"Failed to fetch calendar for {email}: {e}")
            return []

    def get_meeting_by_id(self, email: str, item_id: str) -> Optional[dict]:
        """Get specific meeting by ID"""
        account = self.get_account(email)
        if not account:
            return None

        try:
            item = account.calendar.get(id=item_id)
            return self._parse_calendar_item(item)
        except Exception as e:
            logger.error(f"Failed to get meeting {item_id}: {e}")
            return None

    def create_meeting(
        self,
        organizer_email: str,
        subject: str,
        start: datetime,
        end: datetime,
        attendees: list[str],
        body: str = "",
        location: str = ""
    ) -> Optional[dict]:
        """Create a new meeting"""
        account = self.get_account(organizer_email)
        if not account:
            return None

        try:
            tz = EWSTimeZone.localzone()

            meeting = CalendarItem(
                account=account,
                folder=account.calendar,
                subject=subject,
                body=body,
                start=tz.localize(EWSDateTime.from_datetime(start)),
                end=tz.localize(EWSDateTime.from_datetime(end)),
                location=location,
                required_attendees=[Attendee(mailbox_email=email) for email in attendees]
            )

            meeting.save(send_meeting_invitations='SendToAllAndSaveCopy')
            logger.info(f"Created meeting: {subject}")

            return self._parse_calendar_item(meeting)

        except Exception as e:
            logger.error(f"Failed to create meeting: {e}")
            return None

    def update_meeting(
        self,
        organizer_email: str,
        item_id: str,
        subject: str = None,
        start: datetime = None,
        end: datetime = None,
        attendees: list[str] = None,
        body: str = None
    ) -> Optional[dict]:
        """Update existing meeting"""
        account = self.get_account(organizer_email)
        if not account:
            return None

        try:
            meeting = account.calendar.get(id=item_id)
            tz = EWSTimeZone.localzone()

            if subject:
                meeting.subject = subject
            if body:
                meeting.body = body
            if start:
                meeting.start = tz.localize(EWSDateTime.from_datetime(start))
            if end:
                meeting.end = tz.localize(EWSDateTime.from_datetime(end))
            if attendees:
                meeting.required_attendees = [Attendee(mailbox_email=email) for email in attendees]

            meeting.save(update_fields=['subject', 'body', 'start', 'end', 'required_attendees'],
                        send_meeting_invitations='SendToChangedAndSaveCopy')

            logger.info(f"Updated meeting: {meeting.subject}")
            return self._parse_calendar_item(meeting)

        except Exception as e:
            logger.error(f"Failed to update meeting: {e}")
            return None

    def delete_meeting(self, organizer_email: str, item_id: str) -> bool:
        """Delete/cancel meeting"""
        account = self.get_account(organizer_email)
        if not account:
            return False

        try:
            meeting = account.calendar.get(id=item_id)
            meeting.delete(send_meeting_cancellations='SendToAllAndSaveCopy')
            logger.info(f"Deleted meeting: {item_id}")
            return True

        except Exception as e:
            logger.error(f"Failed to delete meeting: {e}")
            return False

    def get_free_busy(self, emails: list[str], start: datetime, end: datetime) -> dict:
        """Get free/busy info for multiple users"""
        if not self.service_account:
            if not self.connect_service_account():
                return {}

        try:
            from exchangelib import GetUserAvailability
            tz = EWSTimeZone.localzone()

            # This requires exchange impersonation rights
            # Simplified version - just return calendar events
            result = {}
            for email in emails:
                events = self.get_calendar_events(email)
                busy_times = []
                for e in events:
                    if e['start'] and e['end']:
                        event_start = datetime.fromisoformat(e['start'])
                        event_end = datetime.fromisoformat(e['end'])
                        if event_start < end and event_end > start:
                            busy_times.append({
                                'start': e['start'],
                                'end': e['end'],
                                'subject': e['subject']
                            })
                result[email] = busy_times

            return result

        except Exception as e:
            logger.error(f"Failed to get free/busy: {e}")
            return {}

    def find_free_slots(
        self,
        emails: list[str],
        duration_minutes: int,
        start: datetime,
        end: datetime,
        working_hours_start: int = 9,
        working_hours_end: int = 18
    ) -> list[dict]:
        """Find available time slots for all attendees"""
        free_busy = self.get_free_busy(emails, start, end)

        # Merge all busy times
        all_busy = []
        for email, busy_times in free_busy.items():
            for bt in busy_times:
                all_busy.append({
                    'start': datetime.fromisoformat(bt['start']),
                    'end': datetime.fromisoformat(bt['end'])
                })

        # Sort by start time
        all_busy.sort(key=lambda x: x['start'])

        # Find gaps
        free_slots = []
        current = start

        for busy in all_busy:
            # Check if there's a gap before this busy period
            if busy['start'] > current:
                gap_start = current
                gap_end = busy['start']

                # Check working hours and duration
                # Simplified - just add the gap
                if (gap_end - gap_start).total_seconds() >= duration_minutes * 60:
                    free_slots.append({
                        'start': gap_start.isoformat(),
                        'end': gap_end.isoformat()
                    })

            current = max(current, busy['end'])

        # Check remaining time until end
        if (end - current).total_seconds() >= duration_minutes * 60:
            free_slots.append({
                'start': current.isoformat(),
                'end': end.isoformat()
            })

        return free_slots

    def _parse_calendar_item(self, item: CalendarItem) -> dict:
        """Parse calendar item to dict"""
        try:
            attendees = []
            if item.required_attendees:
                for a in item.required_attendees:
                    attendees.append({
                        'email': a.mailbox.email_address if a.mailbox else None,
                        'name': a.mailbox.name if a.mailbox else None,
                        'response': str(a.response_type) if hasattr(a, 'response_type') else None
                    })
            if item.optional_attendees:
                for a in item.optional_attendees:
                    attendees.append({
                        'email': a.mailbox.email_address if a.mailbox else None,
                        'name': a.mailbox.name if a.mailbox else None,
                        'response': str(a.response_type) if hasattr(a, 'response_type') else None,
                        'optional': True
                    })

            return {
                'id': item.id if hasattr(item, 'id') else None,
                'subject': item.subject,
                'body': str(item.body) if item.body else None,
                'start': item.start.isoformat() if item.start else None,
                'end': item.end.isoformat() if item.end else None,
                'location': item.location if hasattr(item, 'location') else None,
                'organizer': item.organizer.email_address if item.organizer else None,
                'attendees': attendees,
                'is_recurring': item.is_recurring if hasattr(item, 'is_recurring') else False,
                'is_cancelled': item.is_cancelled if hasattr(item, 'is_cancelled') else False
            }

        except Exception as e:
            logger.error(f"Failed to parse calendar item: {e}")
            return {}

    def _extract_email(self, username: str) -> str:
        """Extract email from username (domain\\user -> need to look up)"""
        # This is a placeholder - you'd typically look this up
        if '@' in username:
            return username
        # For domain\\user format, you need to know the email domain
        user = username.split('\\')[-1] if '\\' in username else username
        return f"{user}@ekf.su"  # Adjust domain as needed


# Test connection
if __name__ == "__main__":
    import yaml
    import os

    logging.basicConfig(level=logging.DEBUG)

    with open('config.yaml', 'r') as f:
        config = yaml.safe_load(f)

    # Expand env vars
    config['ews']['username'] = os.environ.get('EWS_USERNAME', config['ews']['username'])
    config['ews']['password'] = os.environ.get('EWS_PASSWORD', config['ews']['password'])

    client = EWSClient(config)
    if client.connect_service_account():
        # Test getting calendar
        email = "test@ekf.su"  # Replace with real email
        events = client.get_calendar_events(email)
        print(f"Found {len(events)} events")
        for e in events[:5]:
            print(f"  - {e['subject']} ({e['start']})")

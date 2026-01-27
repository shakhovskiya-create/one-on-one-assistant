import { useEffect, useRef, useCallback, useState } from 'react';
import { getWebSocketUrl } from '@/lib/api/client';
import { useAuthStore } from '@/stores/auth';
import type { Message, Channel } from '@/types';

interface WebSocketMessage {
  type: 'message' | 'typing' | 'presence' | 'channel_update' | 'reaction';
  payload: unknown;
}

interface UseWebSocketOptions {
  onMessage?: (message: Message) => void;
  onTyping?: (data: { channel_id: string; user_id: string }) => void;
  onPresence?: (data: { user_id: string; status: 'online' | 'offline' }) => void;
  onChannelUpdate?: (channel: Channel) => void;
  onReaction?: (data: { message_id: string; emoji: string; user_id: string }) => void;
}

export function useWebSocket(options: UseWebSocketOptions = {}) {
  const { onMessage, onTyping, onPresence, onChannelUpdate, onReaction } = options;
  const user = useAuthStore((state) => state.user);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<number | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const connect = useCallback(() => {
    if (!user?.id || wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    const wsUrl = getWebSocketUrl(user.id);
    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      console.log('WebSocket connected');
      setIsConnected(true);
      setError(null);
    };

    ws.onmessage = (event) => {
      try {
        const data: WebSocketMessage = JSON.parse(event.data);

        switch (data.type) {
          case 'message':
            onMessage?.(data.payload as Message);
            break;
          case 'typing':
            onTyping?.(data.payload as { channel_id: string; user_id: string });
            break;
          case 'presence':
            onPresence?.(data.payload as { user_id: string; status: 'online' | 'offline' });
            break;
          case 'channel_update':
            onChannelUpdate?.(data.payload as Channel);
            break;
          case 'reaction':
            onReaction?.(data.payload as { message_id: string; emoji: string; user_id: string });
            break;
        }
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err);
      }
    };

    ws.onerror = (event) => {
      console.error('WebSocket error:', event);
      setError('Ошибка подключения к серверу');
    };

    ws.onclose = (event) => {
      console.log('WebSocket closed:', event.code, event.reason);
      setIsConnected(false);
      wsRef.current = null;

      // Reconnect after 3 seconds if not intentionally closed
      if (event.code !== 1000) {
        reconnectTimeoutRef.current = window.setTimeout(() => {
          connect();
        }, 3000);
      }
    };

    wsRef.current = ws;
  }, [user?.id, onMessage, onTyping, onPresence, onChannelUpdate, onReaction]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    if (wsRef.current) {
      wsRef.current.close(1000, 'User disconnect');
      wsRef.current = null;
    }
    setIsConnected(false);
  }, []);

  const send = useCallback((type: string, payload: unknown) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type, payload }));
    }
  }, []);

  const sendMessage = useCallback((channelId: string, content: string) => {
    send('message', { channel_id: channelId, content });
  }, [send]);

  const sendTyping = useCallback((channelId: string) => {
    send('typing', { channel_id: channelId });
  }, [send]);

  const sendReaction = useCallback((messageId: string, emoji: string) => {
    send('reaction', { message_id: messageId, emoji });
  }, [send]);

  // Auto-connect when user is logged in
  useEffect(() => {
    if (user?.id) {
      connect();
    } else {
      disconnect();
    }

    return () => {
      disconnect();
    };
  }, [user?.id, connect, disconnect]);

  return {
    isConnected,
    error,
    send,
    sendMessage,
    sendTyping,
    sendReaction,
    connect,
    disconnect,
  };
}

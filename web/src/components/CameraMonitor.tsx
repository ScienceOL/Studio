import { apiClient } from '@/service/http/client';
import { getAuthenticatedWsUrl } from '@/service/ws/client';
import { useEffect, useRef, useState } from 'react';

interface Props {
  hostId: string;
  cameraId: string;
}

interface SignalMessage {
  type: 'offer' | 'answer' | 'ice-candidate';
  hostId?: string;
  cameraId?: string;
  sdp?: string;
  candidate?: RTCIceCandidateInit;
}

// Simplified WebRTC player (prototype) based on demo/WebRTCPlayer.tsx
export const CameraMonitor: React.FC<Props> = ({ hostId, cameraId }) => {
  const videoRef = useRef<HTMLVideoElement>(null);
  const pcRef = useRef<RTCPeerConnection | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  const [streaming, setStreaming] = useState(false);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [stats, setStats] = useState<{ fps: number; latency: number } | null>(
    null
  );
  const statsTimerRef = useRef<number | null>(null);
  const remoteSetRef = useRef(false);

  useEffect(() => {
    // auto start on mount
    start();
    return () => cleanup();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const cleanup = () => {
    if (pcRef.current) {
      pcRef.current.close();
      pcRef.current = null;
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setStreaming(false);
    setConnected(false);
  };

  const connectWS = (): Promise<void> => {
    return new Promise((resolve, reject) => {
      const clientId = `client-${Date.now()}`;
      // Use getAuthenticatedWsUrl to include auth token and handle protocol
      const wsUrl = getAuthenticatedWsUrl(
        `/api/realtime/signal/client?clientId=${clientId}`
      );
      const ws = new WebSocket(wsUrl);

      ws.onopen = () => {
        setConnected(true);
        wsRef.current = ws;
        resolve();
      };
      ws.onerror = (e) => {
        setError('WebSocket error');
        reject(e);
      };
      ws.onclose = () => setConnected(false);
      ws.onmessage = async (ev) => {
        const msg: SignalMessage = JSON.parse(ev.data);
        const pc = pcRef.current;
        if (!pc) return;

        // Ignore messages for other host/camera (since backend broadcasts)
        if (msg.hostId && msg.hostId !== hostId) return;
        if (msg.cameraId && msg.cameraId !== cameraId) return;

        if (msg.type === 'answer' && msg.sdp) {
          // Only accept answer when we are in have-local-offer state
          if (pc.signalingState !== 'have-local-offer') {
            console.warn(
              '[CameraMonitor] Stale answer ignored, state=',
              pc.signalingState
            );
            return;
          }
          try {
            await pc.setRemoteDescription(
              new RTCSessionDescription({ type: 'answer', sdp: msg.sdp })
            );
            remoteSetRef.current = true;
            console.log('[CameraMonitor] Remote SDP applied');
          } catch (err) {
            console.warn('[CameraMonitor] setRemoteDescription error', err);
          }
        } else if (msg.type === 'ice-candidate' && msg.candidate) {
          try {
            await pc.addIceCandidate(new RTCIceCandidate(msg.candidate));
          } catch (err) {
            console.warn('addIceCandidate error', err);
          }
        }
      };
    });
  };

  const sendSignal = (m: SignalMessage) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(m));
    }
  };

  const collectStats = async () => {
    if (!pcRef.current) return;
    try {
      const reports = await pcRef.current.getStats();
      let fps = 0;
      let latency = 0;
      reports.forEach((r) => {
        // Narrow stats shape without using any
        const report = r as unknown as {
          type: string;
          kind?: string;
          framesPerSecond?: number;
          state?: string;
          currentRoundTripTime?: number;
        };
        if (report.type === 'inbound-rtp' && report.kind === 'video') {
          fps = report.framesPerSecond ?? fps;
        }
        if (report.type === 'candidate-pair' && report.state === 'succeeded') {
          latency = (report.currentRoundTripTime ?? 0) * 1000; // ms
        }
      });
      setStats({ fps: Math.round(fps), latency: Math.round(latency) });
    } catch {
      // ignore stats errors in prototype
    }
  };

  const startStatsInterval = () => {
    if (statsTimerRef.current) window.clearInterval(statsTimerRef.current);
    statsTimerRef.current = window.setInterval(collectStats, 1000);
  };

  const start = async () => {
    // Cleanup previous session to avoid state conflicts
    cleanup();

    try {
      setError(null);
      // Ensure WS is connected
      if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
        await connectWS();
      }

      const pc = new RTCPeerConnection({
        iceServers: [{ urls: 'stun:stun.l.google.com:19302' }],
      });
      pcRef.current = pc;

      pc.onicecandidate = (e) => {
        if (e.candidate) {
          sendSignal({
            type: 'ice-candidate',
            hostId,
            cameraId,
            candidate: e.candidate.toJSON(),
          });
        }
      };

      pc.ontrack = (e) => {
        if (videoRef.current && e.streams[0]) {
          videoRef.current.srcObject = e.streams[0];
        }
      };

      pc.addTransceiver('video', { direction: 'recvonly' });
      pc.addTransceiver('audio', { direction: 'recvonly' });

      const offer = await pc.createOffer();
      await pc.setLocalDescription(offer);
      remoteSetRef.current = false;

      sendSignal({ type: 'offer', hostId, cameraId, sdp: offer.sdp });

      // Notify backend to start stream
      // Use apiClient for authenticated request and correct base URL
      await apiClient.post('/api/realtime/camera/start', { hostId, cameraId });

      setStreaming(true);
      startStatsInterval();
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('unknown error');
      }
      cleanup();
    }
  };

  // No manual stop button; component cleans on unmount.

  return (
    <div className="relative w-full max-w-full overflow-hidden rounded-lg bg-black shadow-sm">
      <video
        ref={videoRef}
        autoPlay
        playsInline
        muted
        className="aspect-video w-full select-none object-cover"
      />
      {/* Overlay */}
      <div className="pointer-events-none absolute left-2 top-2 flex flex-col gap-1 rounded bg-black/40 px-3 py-2 text-xs font-medium text-white backdrop-blur-sm">
        <div className="flex items-center gap-2">
          <span className={connected ? 'text-green-400' : 'text-red-400'}>
            {connected ? 'LIVE' : 'OFF'}
          </span>
          <span>{streaming ? 'Streaming' : 'Idle'}</span>
        </div>
        <div className="flex gap-2 opacity-80">
          <span>Ans:{remoteSetRef.current ? 'Y' : 'N'}</span>
        </div>
        {stats && (
          <div className="flex gap-4">
            <span>FPS: {stats.fps}</span>
            <span>Latency: {stats.latency}ms</span>
          </div>
        )}
        {error && <span className="text-red-300">ERR: {error}</span>}
        {!remoteSetRef.current && !error && (
          <span className="text-yellow-300">Waiting answer...</span>
        )}
      </div>
      <button
        type="button"
        onClick={() => start()}
        className="absolute bottom-2 right-2 rounded bg-white/10 px-2 py-1 text-[10px] text-white hover:bg-white/20"
      >
        Reconnect
      </button>
    </div>
  );
};

export default CameraMonitor;

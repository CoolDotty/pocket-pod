import { useEffect, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import type { PodmanContainer } from "@/types/podman";
import { podmanQueryKeys } from "./podmanQueries";

type PodmanStreamMessage =
  | { type: "containers"; data?: PodmanContainer[]; message?: string }
  | { type: "error"; data?: PodmanContainer[]; message?: string };

type PodmanStreamStatus = "idle" | "connecting" | "open" | "closed" | "error";

export const usePodmanContainersStream = (enabled: boolean) => {
  const queryClient = useQueryClient();
  const [streamStatus, setStreamStatus] =
    useState<PodmanStreamStatus>("idle");
  const [streamError, setStreamError] = useState<string | null>(null);
  const socketRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    let isActive = true;
    let reconnectTimer: number | null = null;
    let retryCount = 0;

    if (!enabled) {
      if (socketRef.current) {
        socketRef.current.close();
        socketRef.current = null;
      }
      if (reconnectTimer) {
        window.clearTimeout(reconnectTimer);
        reconnectTimer = null;
      }
      if (isActive) {
        setStreamStatus("idle");
        setStreamError(null);
      }
      return;
    }

    const protocol = window.location.protocol === "https:" ? "wss" : "ws";
    const wsUrl = `${protocol}://${window.location.host}/podman/containers/stream`;
    const scheduleReconnect = () => {
      if (!isActive || reconnectTimer) {
        return;
      }
      const delay = Math.min(10000, 1000 * Math.pow(2, retryCount));
      retryCount += 1;
      reconnectTimer = window.setTimeout(() => {
        reconnectTimer = null;
        connect();
      }, delay);
    };

    const connect = () => {
      if (!isActive) {
        return;
      }

      if (socketRef.current) {
        socketRef.current.close();
      }

      const socket = new WebSocket(wsUrl);
      socketRef.current = socket;

      if (isActive) {
        setStreamStatus("connecting");
        setStreamError(null);
      }

      socket.onopen = () => {
        if (isActive) {
          retryCount = 0;
          setStreamStatus("open");
        }
      };

      socket.onclose = () => {
        if (isActive) {
          setStreamStatus("closed");
          scheduleReconnect();
        }
      };

      socket.onerror = () => {
        if (isActive) {
          setStreamStatus("error");
          setStreamError("Live updates disconnected.");
        }
        socket.close();
      };

      socket.onmessage = (event) => {
        if (typeof event.data !== "string") {
          return;
        }

        let payload: PodmanStreamMessage | null = null;
        try {
          payload = JSON.parse(event.data) as PodmanStreamMessage;
        } catch {
          return;
        }

        if (!payload) {
          return;
        }

        if (payload.type === "containers") {
          const containers = Array.isArray(payload.data) ? payload.data : [];
          queryClient.setQueryData(podmanQueryKeys.containers, containers);
          if (isActive) {
            setStreamError(null);
          }
          return;
        }

        if (payload.type === "error") {
          if (payload.message && payload.message.trim() !== "") {
            if (isActive) {
              setStreamError(payload.message);
            }
          }
        }
      };
    };

    connect();

    return () => {
      isActive = false;
      if (reconnectTimer) {
        window.clearTimeout(reconnectTimer);
        reconnectTimer = null;
      }
      if (socketRef.current) {
        socketRef.current.close();
      }
      socketRef.current = null;
    };
  }, [enabled, queryClient]);

  return { streamStatus, streamError };
};

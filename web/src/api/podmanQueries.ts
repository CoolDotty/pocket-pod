import { isAxiosError } from "axios";
import { useQuery, type UseQueryOptions } from "@tanstack/react-query";
import type { PodmanContainer } from "@/types/podman";
import apiClient from "./client";

export const podmanQueryKeys = {
  containers: ["podman-containers"] as const,
};

const fetchPodmanContainers = async () => {
  const response = await apiClient.get<PodmanContainer[]>("/podman/containers");
  return response.data;
};

export const getPodmanContainersErrorMessage = (err: unknown) => {
  if (isAxiosError(err)) {
    const message = err.response?.data?.message;
    if (typeof message === "string" && message.trim() !== "") {
      return message;
    }
  }
  return "Failed to load Podman containers.";
};

type PodmanQueryOptions<
  TQueryFnData,
  TQueryKey extends readonly unknown[],
> = Omit<
  UseQueryOptions<TQueryFnData, unknown, TQueryFnData, TQueryKey>,
  "queryKey" | "queryFn"
>;

export const usePodmanContainersQuery = (
  options?: PodmanQueryOptions<
    PodmanContainer[],
    typeof podmanQueryKeys.containers
  >,
) =>
  useQuery({
    queryKey: podmanQueryKeys.containers,
    queryFn: fetchPodmanContainers,
    retry: false,
    ...options,
  });

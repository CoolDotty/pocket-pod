export type PodmanContainer = {
  id: string;
  name: string;
  image: string;
  status: string;
  tunnelStatus?: "ready" | "starting" | "blocked" | "failed";
  tunnelCode?: string;
  tunnelMessage?: string;
  tunnelUrl?: string;
  storageSize?: string;
  createdAt?: string;
  ports?: string;
  labels?: Record<string, string>;
};

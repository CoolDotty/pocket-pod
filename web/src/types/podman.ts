export type PodmanContainer = {
  id: string;
  name: string;
  image: string;
  status: string;
  storageSize?: string;
  createdAt?: string;
  ports?: string;
  labels?: Record<string, string>;
};

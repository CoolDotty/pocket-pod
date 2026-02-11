export type PodmanContainer = {
  id: string;
  name: string;
  image: string;
  status: string;
  createdAt?: string;
  ports?: string;
  labels?: Record<string, string>;
};

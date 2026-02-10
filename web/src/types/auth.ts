export type User = {
  id: string;
  email: string;
  role: string;
  display_name: string;
};

export type SignupConfig = {
  requiresInvite: boolean;
  userCount: number;
};

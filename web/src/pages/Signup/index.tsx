import { useMemo } from "react";
import { Navigate, useNavigate } from "react-router-dom";
import LoadingCard from "../../components/LoadingCard";
import type { SignupConfig, User } from "../../types/auth";
import type { FormEvent } from "react";

type SignupPageProps = {
  user: User | null;
  loading: boolean;
  signupConfig: SignupConfig | null;
  error: string | null;
  submitting: boolean;
  signupEmail: string;
  signupPassword: string;
  signupPasswordConfirm: string;
  inviteToken: string;
  onSignup: (event: FormEvent<HTMLFormElement>) => void;
  setSignupEmail: (value: string) => void;
  setSignupPassword: (value: string) => void;
  setSignupPasswordConfirm: (value: string) => void;
  setInviteToken: (value: string) => void;
};

export default function SignupPage({
  user,
  loading,
  signupConfig,
  error,
  submitting,
  signupEmail,
  signupPassword,
  signupPasswordConfirm,
  inviteToken,
  onSignup,
  setSignupEmail,
  setSignupPassword,
  setSignupPasswordConfirm,
  setInviteToken,
}: SignupPageProps) {
  const navigate = useNavigate();
  const requiresInvite = useMemo(
    () => signupConfig?.requiresInvite ?? false,
    [signupConfig],
  );

  if (loading) {
    return <LoadingCard />;
  }

  if (user) {
    return <Navigate to="/dashboard" replace />;
  }

  return (
    <section className="card auth-card">
      <h2>Signup</h2>
      <div className="auth-forms">
        <form className="auth-form" onSubmit={onSignup}>
          <h3>Sign up</h3>
          <label>
            Email
            <input
              type="email"
              value={signupEmail}
              onChange={(event) => setSignupEmail(event.target.value)}
              autoComplete="email"
              required
            />
          </label>
          <label>
            Password
            <input
              type="password"
              value={signupPassword}
              onChange={(event) => setSignupPassword(event.target.value)}
              autoComplete="new-password"
              required
            />
          </label>
          <label>
            Confirm password
            <input
              type="password"
              value={signupPasswordConfirm}
              onChange={(event) =>
                setSignupPasswordConfirm(event.target.value)
              }
              autoComplete="new-password"
              required
            />
          </label>
          {requiresInvite ? (
            <label>
              Invite token
              <input
                type="text"
                value={inviteToken}
                onChange={(event) => setInviteToken(event.target.value)}
                required
              />
            </label>
          ) : (
            <p className="muted">
              First user signup is open. Subsequent signups require an invite
              token.
            </p>
          )}
          <button className="button" type="submit" disabled={submitting}>
            Create account
          </button>
          <button
            className="button outline"
            type="button"
            onClick={() => navigate("/login")}
            disabled={submitting}
          >
            Back to login
          </button>
        </form>
      </div>
      {error ? <p className="error">{error}</p> : null}
    </section>
  );
}

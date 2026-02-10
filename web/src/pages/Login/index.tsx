import { Navigate, useNavigate } from "react-router-dom";
import LoadingCard from "../../components/LoadingCard";
import type { User } from "../../types/auth";
import type { FormEvent } from "react";

type LoginPageProps = {
  user: User | null;
  loading: boolean;
  error: string | null;
  submitting: boolean;
  loginEmail: string;
  loginPassword: string;
  onLogin: (event: FormEvent<HTMLFormElement>) => void;
  setLoginEmail: (value: string) => void;
  setLoginPassword: (value: string) => void;
};

export default function LoginPage({
  user,
  loading,
  error,
  submitting,
  loginEmail,
  loginPassword,
  onLogin,
  setLoginEmail,
  setLoginPassword,
}: LoginPageProps) {
  const navigate = useNavigate();

  if (loading) {
    return <LoadingCard />;
  }

  if (user) {
    return <Navigate to="/dashboard" replace />;
  }

  return (
    <section className="card auth-card">
      <h2>Login</h2>
      <div className="auth-forms">
        <form className="auth-form" onSubmit={onLogin}>
          <h3>Log in</h3>
          <label>
            Email
            <input
              type="email"
              value={loginEmail}
              onChange={(event) => setLoginEmail(event.target.value)}
              autoComplete="email"
              required
            />
          </label>
          <label>
            Password
            <input
              type="password"
              value={loginPassword}
              onChange={(event) => setLoginPassword(event.target.value)}
              autoComplete="current-password"
              required
            />
          </label>
          <button className="button" type="submit" disabled={submitting}>
            Log in
          </button>
          <button
            className="button outline"
            type="button"
            onClick={() => navigate("/signup")}
            disabled={submitting}
          >
            Create account
          </button>
        </form>
      </div>
      {error ? <p className="error">{error}</p> : null}
    </section>
  );
}

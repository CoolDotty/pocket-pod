import { Navigate } from "react-router-dom";
import LoadingCard from "../../components/LoadingCard";
import type { User } from "../../types/auth";

type DashboardPageProps = {
  user: User | null;
  loading: boolean;
  submitting: boolean;
  error: string | null;
  onLogout: () => void;
};

export default function DashboardPage({
  user,
  loading,
  submitting,
  error,
  onLogout,
}: DashboardPageProps) {
  if (loading) {
    return <LoadingCard />;
  }

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  return (
    <section className="card auth-card">
      <h2>Dashboard</h2>
      <p className="muted">You are signed in and ready to go.</p>
      <div className="auth-details" style={{ marginTop: "1rem" }}>
        <div>
          <strong>{user.display_name || user.email}</strong>
          <div className="muted">{user.email}</div>
        </div>
        <span className="pill">{user.role}</span>
      </div>
      <div className="auth-forms" style={{ marginTop: "1.5rem" }}>
        <button
          className="button outline"
          type="button"
          onClick={onLogout}
          disabled={submitting}
        >
          Log out
        </button>
      </div>
      {error ? <p className="error">{error}</p> : null}
    </section>
  );
}

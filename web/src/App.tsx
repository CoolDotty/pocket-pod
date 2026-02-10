import { useEffect, useState } from "react";
import type { FormEvent } from "react";
import { Navigate, Route, Routes, useLocation, useNavigate } from "react-router-dom";
import DashboardPage from "./pages/Dashboard";
import LoginPage from "./pages/Login";
import SignupPage from "./pages/Signup";
import { fetchJson } from "./lib/fetchJson";
import type { SignupConfig, User } from "./types/auth";

export default function App() {
  const [user, setUser] = useState<User | null>(null);
  const [signupConfig, setSignupConfig] = useState<SignupConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [loginEmail, setLoginEmail] = useState("");
  const [loginPassword, setLoginPassword] = useState("");
  const [signupEmail, setSignupEmail] = useState("");
  const [signupPassword, setSignupPassword] = useState("");
  const [signupPasswordConfirm, setSignupPasswordConfirm] = useState("");
  const [inviteToken, setInviteToken] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const navigate = useNavigate();
  const location = useLocation();

  const loadSession = async () => {
    setLoading(true);
    setError(null);

    const [meResult, configResult] = await Promise.all([
      fetch("/auth/me", { credentials: "include" }),
      fetch("/auth/signup-config", { credentials: "include" }),
    ]);

    if (meResult.ok) {
      setUser((await meResult.json()) as User);
    } else {
      setUser(null);
    }

    if (configResult.ok) {
      setSignupConfig((await configResult.json()) as SignupConfig);
    }

    setLoading(false);
  };

  useEffect(() => {
    void loadSession();
  }, []);

  useEffect(() => {
    setError(null);
  }, [location.pathname]);

  const handleLogin = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setSubmitting(true);
    setError(null);

    try {
      const data = await fetchJson<User>("/auth/login", {
        method: "POST",
        body: JSON.stringify({ email: loginEmail, password: loginPassword }),
      });
      setUser(data);
      setLoginPassword("");
      navigate("/dashboard", { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setSubmitting(false);
    }
  };

  const handleSignup = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setSubmitting(true);
    setError(null);

    try {
      const data = await fetchJson<User>("/auth/signup", {
        method: "POST",
        body: JSON.stringify({
          email: signupEmail,
          password: signupPassword,
          passwordConfirm: signupPasswordConfirm,
          inviteToken: inviteToken || undefined,
        }),
      });
      setUser(data);
      setSignupPassword("");
      setSignupPasswordConfirm("");
      setInviteToken("");
      navigate("/dashboard", { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Signup failed");
    } finally {
      setSubmitting(false);
    }
  };

  const handleLogout = async () => {
    setSubmitting(true);
    setError(null);

    try {
      await fetchJson<void>("/auth/logout", { method: "POST" });
      setUser(null);
      navigate("/login", { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Logout failed");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="app">
      <header className="hero">
        <p className="eyebrow">PocketBase + Go + React</p>
        <h1>Urban Octo Umbrella</h1>
        <p className="lede">
          Single-binary deployment with PocketBase embedded in Go and a React
          frontend served from the same executable.
        </p>
      </header>

      <Routes>
        <Route
          path="/"
          element={<Navigate to={user ? "/dashboard" : "/login"} replace />}
        />
        <Route
          path="/dashboard"
          element={
            <DashboardPage
              user={user}
              loading={loading}
              submitting={submitting}
              error={error}
              onLogout={handleLogout}
            />
          }
        />
        <Route
          path="/login"
          element={
            <LoginPage
              user={user}
              loading={loading}
              error={error}
              submitting={submitting}
              loginEmail={loginEmail}
              loginPassword={loginPassword}
              onLogin={handleLogin}
              setLoginEmail={setLoginEmail}
              setLoginPassword={setLoginPassword}
            />
          }
        />
        <Route
          path="/signup"
          element={
            <SignupPage
              user={user}
              loading={loading}
              signupConfig={signupConfig}
              error={error}
              submitting={submitting}
              signupEmail={signupEmail}
              signupPassword={signupPassword}
              signupPasswordConfirm={signupPasswordConfirm}
              inviteToken={inviteToken}
              onSignup={handleSignup}
              setSignupEmail={setSignupEmail}
              setSignupPassword={setSignupPassword}
              setSignupPasswordConfirm={setSignupPasswordConfirm}
              setInviteToken={setInviteToken}
            />
          }
        />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </div>
  );
}

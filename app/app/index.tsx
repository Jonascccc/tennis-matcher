// app/app/index.tsx
import { Redirect } from "expo-router";
import { useAuth } from "../src/auth";
export default function Index() {
  const { token, loading } = useAuth();
  if (loading) return null;
  return <Redirect href={token ? "/match" : "/login"} />;
}

// app/app/_layout.tsx
import { Stack } from "expo-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthProvider } from "../src/auth";
const qc = new QueryClient();
export default function Root() {
  return (
    <QueryClientProvider client={qc}>
      <AuthProvider>
        <Stack screenOptions={{ headerTitle: "Tennis" }}>
          <Stack.Screen name="index" options={{ headerShown:false }} />
          <Stack.Screen name="login" />
          <Stack.Screen name="register" />
          <Stack.Screen name="profile" />
          <Stack.Screen name="match" />
        </Stack>
      </AuthProvider>
    </QueryClientProvider>
  );
}

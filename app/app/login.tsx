// app/app/login.tsx
import React, { useState } from "react";
import { View, TextInput, Button, Text, Alert } from "react-native";
import { Link, useRouter } from "expo-router";
import * as WebBrowser from "expo-web-browser";
import * as Google from "expo-auth-session/providers/google";
import { useLogin, useLoginGoogle } from "../src/hooks";

WebBrowser.maybeCompleteAuthSession();

export default function Login() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const { mutateAsync, isPending } = useLogin();
  const loginGoogle = useLoginGoogle();
  const router = useRouter();

  // Use EXPO_PUBLIC_  environment variables
  const [_, response, promptAsync] = Google.useIdTokenAuthRequest({
    clientId: process.env.EXPO_PUBLIC_GOOGLE_EXPO_CLIENT_ID,
    androidClientId: process.env.EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID,
    iosClientId: process.env.EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID,
  });

  const onSubmit = async () => {
    try {
      await mutateAsync({ email, password });
      router.replace("/match");
    } catch (e: any) {
      Alert.alert("Login failed", e?.response?.data?.error || "unknown");
    }
  };

  const onGoogle = async () => {
    try {
      const res = await promptAsync();
      if (res?.type === "success" && (res as any).params?.id_token) {
        await loginGoogle.mutateAsync((res as any).params.id_token);
        router.replace("/match");
      }
    } catch (e: any) {
      Alert.alert("Google login failed", e?.response?.data?.error || e?.message || "unknown");
    }
  };

  return (
    <View style={{ padding: 16, gap: 12 }}>
      <Text style={{ fontSize: 22, fontWeight: "600" }}>Login</Text>
      <TextInput
        placeholder="Email"
        autoCapitalize="none"
        value={email}
        onChangeText={setEmail}
        style={{ borderWidth: 1, padding: 8 }}
      />
      <TextInput
        placeholder="Password"
        secureTextEntry
        value={password}
        onChangeText={setPassword}
        style={{ borderWidth: 1, padding: 8 }}
      />
      <Button title={isPending ? "..." : "Email Password Login"} onPress={onSubmit} />
      <Button title={loginGoogle.isPending ? "..." : "Use Google Login"} onPress={onGoogle} />
      <Text>
        No account?<Link href="/register">Register</Link>
      </Text>
    </View>
  );
}

// app/app/login.tsx
import React, { useState } from "react";
import { View, TextInput, Button, Text, Alert } from "react-native";
import { Link, useRouter } from "expo-router";
import { useLogin } from "../src/hooks";
export default function Login() {
  const [email, setEmail] = useState(""); const [password, setPassword] = useState("");
  const { mutateAsync, isPending } = useLogin(); const router = useRouter();
  const onSubmit = async () => {
    try { await mutateAsync({email,password}); router.replace("/match"); }
    catch (e:any){ Alert.alert("登录失败", e?.response?.data?.error || "unknown"); }
  };
  return (
    <View style={{ padding:16, gap:12 }}>
      <Text style={{ fontSize:22, fontWeight:"600" }}>登录</Text>
      <TextInput placeholder="Email" autoCapitalize="none" value={email} onChangeText={setEmail} style={{ borderWidth:1, padding:8 }} />
      <TextInput placeholder="Password" secureTextEntry value={password} onChangeText={setPassword} style={{ borderWidth:1, padding:8 }} />
      <Button title={isPending?"...":"登录"} onPress={onSubmit} />
      <Text>没有账号？<Link href="/register">去注册</Link></Text>
    </View>
  );
}

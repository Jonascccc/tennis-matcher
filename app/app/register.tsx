// app/app/register.tsx
import React, { useState } from "react";
import { View, TextInput, Button, Text, Alert } from "react-native";
import { Link, useRouter } from "expo-router";
import { useRegister } from "../src/hooks";
export default function Register() {
  const [email,setEmail]=useState(""); const [password,setPassword]=useState("");
  const { mutateAsync, isPending } = useRegister(); const router = useRouter();
  const onSubmit = async () => {
    try { await mutateAsync({email,password}); router.replace("/profile"); }
    catch (e:any){ Alert.alert("注册失败", e?.response?.data?.error || "unknown"); }
  };
  return (
    <View style={{ padding:16, gap:12 }}>
      <Text style={{ fontSize:22, fontWeight:"600" }}>注册</Text>
      <TextInput placeholder="Email" autoCapitalize="none" value={email} onChangeText={setEmail} style={{ borderWidth:1, padding:8 }} />
      <TextInput placeholder="Password" secureTextEntry value={password} onChangeText={setPassword} style={{ borderWidth:1, padding:8 }} />
      <Button title={isPending?"...":"注册"} onPress={onSubmit} />
      <Text>已有账号？<Link href="/login">去登录</Link></Text>
    </View>
  );
}

// app/app/profile.tsx
import React, { useEffect, useState } from "react";
import { View, Text, TextInput, Button, Alert, Switch } from "react-native";
import { useProfile, useSaveProfile } from "../src/hooks";
export default function Profile() {
  const { data, isLoading } = useProfile();
  const save = useSaveProfile();
  const [handedness, setHand] = useState<"R"|"L">("R");
  const [levelEst, setLevel] = useState("2.0");
  const [radiusKm, setRadius] = useState("10");
  const [singles, setSingles] = useState(true);
  const [doubles, setDoubles] = useState(false);
  const [homeLat, setLat] = useState(""); const [homeLng, setLng] = useState("");
  useEffect(() => {
    if (!data) return;
    setHand((data.handedness as any) || "R");
    setLevel(String(data.levelEst ?? 2.0));
    setRadius(String(data.radiusKm ?? 10));
    setSingles((data.preferredFormats || ["SINGLES"]).includes("SINGLES"));
    setDoubles((data.preferredFormats || []).includes("DOUBLES"));
    setLat(data.homeLat != null ? String(data.homeLat) : "");
    setLng(data.homeLng != null ? String(data.homeLng) : "");
  }, [data]);
  if (isLoading) return null;
  return (
    <View style={{ padding:16, gap:12 }}>
      <Text style={{ fontSize:22, fontWeight:"600" }}>我的资料</Text>
      <Text>持拍手 (R/L)</Text>
      <TextInput value={handedness} onChangeText={(t)=>setHand(t==="L"?"L":"R")} style={{ borderWidth:1, padding:8 }} />
      <Text>水平估计 levelEst</Text>
      <TextInput keyboardType="decimal-pad" value={levelEst} onChangeText={setLevel} style={{ borderWidth:1, padding:8 }} />
      <Text>寻找半径（km）</Text>
      <TextInput keyboardType="numeric" value={radiusKm} onChangeText={setRadius} style={{ borderWidth:1, padding:8 }} />
      <Text>偏好赛制</Text>
      <View style={{ flexDirection:"row", alignItems:"center", gap:12 }}>
        <Text>SINGLES</Text><Switch value={singles} onValueChange={setSingles}/>
        <Text>DOUBLES</Text><Switch value={doubles} onValueChange={setDoubles}/>
      </View>
      <Text>家地址 (lat, lng)</Text>
      <View style={{ flexDirection:"row", gap:8 }}>
        <TextInput placeholder="lat" keyboardType="decimal-pad" value={homeLat} onChangeText={setLat} style={{ flex:1, borderWidth:1, padding:8 }} />
        <TextInput placeholder="lng" keyboardType="decimal-pad" value={homeLng} onChangeText={setLng} style={{ flex:1, borderWidth:1, padding:8 }} />
      </View>
      <Button
        title={save.isPending ? "保存中..." : "保存"}
        onPress={async () => {
          try {
            await save.mutateAsync({
              handedness,
              levelEst: parseFloat(levelEst || "0"),
              preferredFormats: [
                ...(singles?["SINGLES"]:[]),
                ...(doubles?["DOUBLES"]:[])
              ],
              radiusKm: parseInt(radiusKm || "0", 10),
              homeLat: homeLat ? parseFloat(homeLat) : undefined,
              homeLng: homeLng ? parseFloat(homeLng) : undefined,
              availability: {},
            });
            Alert.alert("已保存");
          } catch (e:any) {
            Alert.alert("保存失败", e?.response?.data?.error || "unknown");
          }
        }}
      />
    </View>
  );
}

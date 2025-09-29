// app/app/match.tsx
import React, { useState } from "react";
import { View, Text, Button, TextInput, Alert, FlatList } from "react-native";
import dayjs from "dayjs";
import { useFindMatches, getCurrentLatLng } from "../src/hooks";
import type { Candidate, Suggestion } from "../src/api";

export default function Match() {
  const [format, setFormat] = useState<"SINGLES"|"DOUBLES">("SINGLES");
  const [radiusKm, setRadius] = useState("10");
  const [center, setCenter] = useState<{lat:number;lng:number}|null>(null);
  const [when, setWhen] = useState(dayjs().toISOString());
  const finder = useFindMatches();
  const [cands, setCands] = useState<Candidate[]>([]);
  const [sugs, setSugs] = useState<Suggestion[]>([]);

  const run = async () => {
    try {
      const c = center ?? (await getCurrentLatLng());
      const data = await finder.mutateAsync({
        centerLat: c.lat, centerLng: c.lng,
        radiusKm: parseFloat(radiusKm || "10"),
        format, windowStart: when, limit: 50,
      });
      setCenter(c);
      setCands(data.candidates);
      setSugs(data.suggestions);
    } catch (e:any) {
      Alert.alert("查找失败", e?.response?.data?.error || e?.message || "unknown");
    }
  };

  return (
    <View style={{ padding:16, gap:12 }}>
      <Text style={{ fontSize:22, fontWeight:"600" }}>寻找对手</Text>
      <Text>赛制（SINGLES/DOUBLES）</Text>
      <TextInput value={format} onChangeText={(t)=>setFormat(t==="DOUBLES"?"DOUBLES":"SINGLES")} style={{ borderWidth:1, padding:8 }} />
      <Text>半径（km）</Text>
      <TextInput keyboardType="numeric" value={radiusKm} onChangeText={setRadius} style={{ borderWidth:1, padding:8 }} />
      <Text>时间窗口开始（RFC3339）</Text>
      <TextInput value={when} onChangeText={setWhen} style={{ borderWidth:1, padding:8 }} />

      <Button title={finder.isPending ? "搜索中..." : "开始匹配"} onPress={run} />
      {!!center && <Text>中心点：{center.lat.toFixed(5)}, {center.lng.toFixed(5)}</Text>}

      <Text style={{ marginTop:12, fontWeight:"600" }}>候选 ({cands.length})</Text>
      <FlatList data={cands} keyExtractor={(i)=>String(i.userId)}
        renderItem={({item}) => (
          <View style={{ paddingVertical:8, borderBottomWidth:1 }}>
            <Text>UID: {item.userId} · Elo: {item.elo}</Text>
            <Text>距离: {(item.meters/1000).toFixed(2)}km · 匹配分: {item.score.toFixed(2)}</Text>
          </View>
        )}
      />

      <Text style={{ marginTop:12, fontWeight:"600" }}>建议时间/球场 ({sugs.length})</Text>
      <FlatList data={sugs} keyExtractor={(_,i)=>String(i)}
        renderItem={({item}) => (
          <View style={{ paddingVertical:8, borderBottomWidth:1 }}>
            <Text>建议：{item.message}</Text>
            <Text>球场ID: {item.courtId}</Text>
            <Text>{dayjs(item.startAt).format("MM/DD HH:mm")} - {dayjs(item.endAt).format("HH:mm")}</Text>
          </View>
        )}
      />
    </View>
  );
}

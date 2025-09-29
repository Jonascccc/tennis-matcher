// app/src/api.ts
import axios, { AxiosError } from "axios";
import * as SecureStore from "expo-secure-store";
import { Platform } from "react-native";
import Constants from "expo-constants";

export type Profile = {
  handedness?: "R" | "L";
  levelEst?: number;
  elo?: number;
  preferredFormats?: string[];
  radiusKm?: number;
  availability?: Record<string, any>;
  homeLat?: number | null;
  homeLng?: number | null;
};

export type FindReq = {
  centerLat: number;
  centerLng: number;
  radiusKm: number;
  format: "SINGLES" | "DOUBLES";
  windowStart?: string; // RFC3339
  limit?: number;
};
export type Candidate = { userId: number; elo: number; meters: number; score: number };
export type Suggestion = { courtId: number; startAt: string; endAt: string; message: string };
export type FindResp = { candidates: Candidate[]; suggestions: Suggestion[] };

function inferBaseURL() {
  const env = process.env.EXPO_PUBLIC_API_BASE;
  if (env) return env;

  const host =
    (Constants as any).expoConfig?.hostUri ||
    (Constants as any).manifest2?.extra?.expoClient?.hostUri;
  if (host && Platform.OS !== "web") {
    const ip = host.split(":")[0];
    return `http://${ip}:8080/api`;
  }
  return "http://localhost:8080/api";
}

export const api = axios.create({ baseURL: inferBaseURL(), timeout: 10000 });

const TOKEN_KEY = "token";
export async function getToken() {
  if (Platform.OS === "web") return localStorage.getItem(TOKEN_KEY) || null;
  return (await SecureStore.getItemAsync(TOKEN_KEY)) || null;
}
export async function setToken(tok: string | null) {
  if (Platform.OS === "web") {
    if (!tok) localStorage.removeItem(TOKEN_KEY);
    else localStorage.setItem(TOKEN_KEY, tok);
  } else {
    if (!tok) await SecureStore.deleteItemAsync(TOKEN_KEY);
    else await SecureStore.setItemAsync(TOKEN_KEY, tok);
  }
}
api.interceptors.request.use(async (cfg) => {
  const token = await getToken();
  if (token) cfg.headers.Authorization = `Bearer ${token}`;
  return cfg;
});
type OnUnauthorized = () => void;
let on401: OnUnauthorized | null = null;
export function register401Handler(cb: OnUnauthorized) { on401 = cb; }
api.interceptors.response.use(
  r => r,
  (e: AxiosError<any>) => {
    if (e.response?.status === 401 && on401) on401();
    return Promise.reject(e);
  }
);

// REST
export async function register(email: string, password: string) {
  const { data } = await api.post<{ token: string }>("/auth/register", { email, password });
  return data.token;
}
export async function login(email: string, password: string) {
  const { data } = await api.post<{ token: string }>("/auth/login", { email, password });
  return data.token;
}
export async function getProfile() {
  const { data } = await api.get<Profile>("/me/profile");
  return data;
}
export async function putProfile(p: Profile) {
  const { data } = await api.put("/me/profile", p);
  return data;
}
export async function matchFind(body: FindReq) {
  const { data } = await api.post<FindResp>("/match/find", body);
  return data;
}

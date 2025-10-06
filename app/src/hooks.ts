// app/src/hooks.ts
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import * as api from "./api";
import * as Location from "expo-location";

export function useLogin() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({email,password}:{email:string;password:string}) => api.login(email,password),
    onSuccess: async (token) => { await api.setToken(token); qc.clear(); },
  });
}
export function useLoginGoogle() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (idToken: string) => api.loginGoogle(idToken),
    onSuccess: async (token) => { await api.setToken(token); qc.clear(); },
  });
}
export function useRegister() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({email,password}:{email:string;password:string}) => api.register(email,password),
    onSuccess: async (token) => { await api.setToken(token); qc.clear(); },
  });
}
export function useProfile() {
  return useQuery({ queryKey:["profile"], queryFn: api.getProfile });
}
export function useSaveProfile() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (p: api.Profile) => api.putProfile(p),
    onSuccess: () => qc.invalidateQueries({ queryKey:["profile"] }),
  });
}
export async function getCurrentLatLng() {
  const { status } = await Location.requestForegroundPermissionsAsync();
  if (status !== "granted") throw new Error("Location permission denied");
  const loc = await Location.getCurrentPositionAsync({});
  return { lat: loc.coords.latitude, lng: loc.coords.longitude };
}
export function useFindMatches() {
  return useMutation({ mutationFn: (b: api.FindReq) => api.matchFind(b) });
}

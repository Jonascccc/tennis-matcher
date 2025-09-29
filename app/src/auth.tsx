// app/src/auth.tsx
import React, { createContext, useContext, useEffect, useState } from "react";
import { getToken, setToken, register401Handler } from "./api";

type Ctx = { token: string|null; loading: boolean; signin: (t:string)=>Promise<void>; signout: ()=>Promise<void> };
const C = createContext<Ctx>({ token:null, loading:true, signin:async()=>{}, signout:async()=>{} });

export function AuthProvider({children}:{children:React.ReactNode}) {
  const [token, setTok] = useState<string|null>(null);
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    getToken().then(t => { setTok(t); setLoading(false); });
    register401Handler(() => { setTok(null); setToken(null); });
  }, []);
  const signin = async (t:string) => { await setToken(t); setTok(t); };
  const signout = async () => { await setToken(null); setTok(null); };
  return <C.Provider value={{ token, loading, signin, signout }}>{children}</C.Provider>;
}
export const useAuth = () => useContext(C);

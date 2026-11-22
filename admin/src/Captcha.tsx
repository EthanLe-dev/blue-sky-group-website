import {useEffect,useRef} from 'react'
type Turnstile={render:(element:HTMLElement,options:Record<string,unknown>)=>string;remove:(id:string)=>void}
declare global {interface Window {turnstile?:Turnstile}}
let ready:Promise<void>|null=null
export default function Captcha({onToken,action='login'}:{onToken:(token:string)=>void;action?:string}){
 const ref=useRef<HTMLDivElement>(null)
 useEffect(()=>{
  if((import.meta.env.DEV || import.meta.env.VITE_TURNSTILE_MODE === 'mock')){onToken('local-test-token');return}
  const key=import.meta.env.VITE_TURNSTILE_SITE_KEY
  if(!key)return
  if(!ready)ready=new Promise((resolve,reject)=>{const script=document.createElement('script');script.src='https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit';script.async=true;script.onload=()=>resolve();script.onerror=reject;document.head.append(script)})
  let active=true;let widget:string|undefined
  ready.then(()=>{if(active&&ref.current&&window.turnstile)widget=window.turnstile.render(ref.current,{sitekey:key,action,callback:onToken,'expired-callback':()=>onToken('')})}).catch(()=>onToken(''))
  return()=>{active=false;if(widget)window.turnstile?.remove(widget)}
 },[onToken,action])
 return <div ref={ref}>{!(import.meta.env.DEV || import.meta.env.VITE_TURNSTILE_MODE === 'mock')&&!import.meta.env.VITE_TURNSTILE_SITE_KEY&&<p>Cần cấu hình Turnstile để đăng nhập công khai.</p>}</div>
}

"use client";
import {useRef} from 'react';
import {useGSAP} from '@gsap/react';
import gsap from 'gsap';
import Link from 'next/link';
gsap.registerPlugin(useGSAP);
export default function Hero(){const scope=useRef<HTMLElement>(null);useGSAP(()=>{const mm=gsap.matchMedia();mm.add('(prefers-reduced-motion: no-preference)',()=>{gsap.from('[data-enter-text]',{y:22,opacity:0,duration:.7,stagger:.09,ease:'power3.out'});gsap.from('.hero-photo',{scale:1.04,duration:1.1,ease:'power2.out'})});return()=>mm.revert()},{scope});return <section ref={scope} className="hero hero--photo" data-enter=""><img className="hero-photo" src="/images/vinh-hy.jpg" width="1920" height="1280" alt="Vịnh Vĩnh Hy với biển xanh, thuyền neo và núi ven bờ" fetchPriority="high"/><div className="wrap"><div className="hero-copy"><p className="eyebrow" data-enter-text="">BLUE SKY GROUP · DU LỊCH & TRẢI NGHIỆM</p><h1 data-enter-text="">Tour Ninh Chữ.<br/>Hai ngày bên biển.</h1><p data-enter-text="">Từ làng gốm Bàu Trúc đến bờ biển Mỹ Hòa. Xem lịch khởi hành và chọn chuyến đi vừa với bạn.</p><div data-enter-text=""><Link className="btn" href="/tours">Tìm tour ↗</Link><Link className="text-action" href="/tours/ninh-chu-2-ngay-1-dem">Xem hành trình Ninh Chữ</Link></div><p className="hero-fact" data-enter-text="">2 ngày 1 đêm · Giá và lịch demo có thể điều chỉnh</p></div></div></section>}

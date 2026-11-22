"use client";
export default function ErrorPage({reset}:{reset:()=>void}){return <main id="main" className="page wrap state"><h1>Chưa tải được nội dung</h1><p>Kết nối có thể bị gián đoạn. Vui lòng thử lại.</p><button className="btn" onClick={reset}>Thử tải lại</button></main>}

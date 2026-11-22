import type {MetadataRoute} from 'next';
import {read,site,type Tour,type Content} from '@/lib/catalog';
export const dynamic='force-dynamic';
export default async function sitemap():Promise<MetadataRoute.Sitemap>{const paths=['','/tours','/about','/contact','/schools','/companies','/blog','/destinations','/faq','/gallery'];for(const[resource,prefix]of [['tours','tours'],['posts','blog'],['destinations','destinations'],['pages','policies']]){let page=1;for(;;){const data=await read<(Tour|Content)[]>('/'+resource+'?page_size=100&page='+page);paths.push(...data.filter(x=>resource!=='pages'||['privacy','terms','payment','cancellation','preparation'].includes(x.slug)).map(x=>'/'+prefix+'/'+x.slug));if(data.length<100)break;page++}}return [...new Set(paths)].map(path=>({url:site+path}))}

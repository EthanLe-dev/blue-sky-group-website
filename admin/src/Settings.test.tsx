// @vitest-environment jsdom
import {afterEach,expect,test,vi} from 'vitest'
import {cleanup,render,screen,fireEvent,waitFor} from '@testing-library/react'
import {authorized} from './lib/session'
import Settings from './Settings'
vi.mock('./lib/session',()=>({authorized:vi.fn()}))
afterEach(()=>{cleanup();vi.resetAllMocks()})
test('settings expose editable defaults and send optimistic version',async()=>{
 const setting={key:'booking_defaults',version:4,value:{hold_minutes:20,deposit_bps:3000}}
 vi.mocked(authorized).mockResolvedValueOnce(setting).mockResolvedValueOnce({...setting,version:5})
 render(<Settings/>);const input=await screen.findByLabelText('Thời gian giữ chỗ (phút)')
 fireEvent.change(input,{target:{value:'22'}});fireEvent.click(screen.getByRole('button',{name:'Lưu cấu hình'}))
 await waitFor(()=>expect(screen.getByRole('status').textContent).toContain('Đã lưu'))
 const args=vi.mocked(authorized).mock.calls[1];expect(JSON.parse(String(args[1]?.body))).toEqual({expected_version:4,value:{hold_minutes:22,deposit_bps:3000}})
})
test('failed settings save preserves edited value and shows recovery',async()=>{
 vi.mocked(authorized).mockResolvedValueOnce({key:'booking_defaults',version:1,value:{hold_minutes:20}}).mockRejectedValueOnce(new Error('conflict'))
 render(<Settings/>);const input=await screen.findByLabelText('Thời gian giữ chỗ (phút)');fireEvent.change(input,{target:{value:'23'}});fireEvent.click(screen.getByRole('button',{name:'Lưu cấu hình'}))
 await waitFor(()=>expect(screen.getByRole('status').textContent).toContain('Không lưu được'))
 expect((input as HTMLInputElement).value).toBe('23')
})

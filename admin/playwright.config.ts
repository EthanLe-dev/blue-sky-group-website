import {defineConfig} from '@playwright/test'
export default defineConfig({testDir:'e2e',workers:1,use:{baseURL:'http://localhost:5173',channel:'chromium',trace:'off',screenshot:'only-on-failure'},reporter:'list'})

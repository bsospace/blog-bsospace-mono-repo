// Utility for counting OpenAI tokens and checking limits in the frontend
// Requires: npm install js-tiktoken
// Usage:
//   import { countTokens, isTokenLimitExceeded } from '@/utils/token';
//   const tokens = await countTokens('your text');
//   const isOver = await isTokenLimitExceeded('your text', 4096);

import { Tiktoken } from 'js-tiktoken/lite';
import cl100k_base from 'js-tiktoken/ranks/cl100k_base';

let encoder: Tiktoken | null = null;

async function getEncoder(): Promise<Tiktoken> {
  if (!encoder) {
    encoder = new Tiktoken(cl100k_base);
  }
  return encoder;
}

export async function countTokens(text: string): Promise<number> {
  const enc = await getEncoder();
  return enc.encode(text).length;
}

export async function isTokenLimitExceeded(text: string, maxTokens: number): Promise<boolean> {
  const tokens = await countTokens(text);
  return tokens > maxTokens;
} 
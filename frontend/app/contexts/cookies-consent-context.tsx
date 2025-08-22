'use client'

import React, { createContext, useContext, ReactNode } from 'react';
import { useCookiesConsent, CookiesConsent, CookiesPreferences } from '@/hooks/use-cookies-consent';

interface CookiesConsentContextType {
  consent: CookiesConsent;
  preferences: CookiesPreferences;
  isLoaded: boolean;
  acceptAll: () => void;
  rejectNonEssential: () => void;
  updatePreferences: (preferences: Partial<CookiesPreferences>) => void;
  clearConsent: () => void;
  hasConsented: () => boolean;
  isAnalyticsEnabled: () => boolean;
  isMarketingEnabled: () => boolean;
}

const CookiesConsentContext = createContext<CookiesConsentContextType | undefined>(undefined);

interface CookiesConsentProviderProps {
  children: ReactNode;
}

export const CookiesConsentProvider: React.FC<CookiesConsentProviderProps> = ({ children }) => {
  const cookiesConsent = useCookiesConsent();

  return (
    <CookiesConsentContext.Provider value={cookiesConsent}>
      {children}
    </CookiesConsentContext.Provider>
  );
};

export const useCookiesConsentContext = (): CookiesConsentContextType => {
  const context = useContext(CookiesConsentContext);
  if (context === undefined) {
    throw new Error('useCookiesConsentContext must be used within a CookiesConsentProvider');
  }
  return context;
};

export default CookiesConsentProvider;

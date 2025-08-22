import { useState, useEffect, useCallback } from 'react';

export type CookiesConsent = 'accepted' | 'rejected' | 'customized' | null;

export interface CookiesPreferences {
  essential: boolean;
  analytics: boolean;
  marketing: boolean;
}

export const useCookiesConsent = () => {
  const [consent, setConsent] = useState<CookiesConsent>(null);
  const [preferences, setPreferences] = useState<CookiesPreferences>({
    essential: true, // Always true
    analytics: false,
    marketing: false,
  });
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    // Load saved preferences from localStorage
    const savedConsent = localStorage.getItem('cookies-consent');
    const savedPreferences = localStorage.getItem('cookies-preferences');
    
    if (savedConsent) {
      setConsent(savedConsent as CookiesConsent);
    }
    
    if (savedPreferences) {
      try {
        const parsed = JSON.parse(savedPreferences);
        setPreferences({
          essential: true, // Always true
          analytics: parsed.analytics || false,
          marketing: parsed.marketing || false,
        });
      } catch (error) {
        console.error('Error parsing cookies preferences:', error);
      }
    }
    
    setIsLoaded(true);
  }, []);

  const acceptAll = useCallback(() => {
    const newPreferences = {
      essential: true,
      analytics: true,
      marketing: true,
    };
    
    setConsent('accepted');
    setPreferences(newPreferences);
    
    localStorage.setItem('cookies-consent', 'accepted');
    localStorage.setItem('cookies-preferences', JSON.stringify(newPreferences));
    
    // Trigger analytics and marketing cookies
    if (typeof window !== 'undefined') {
      // Enable Google Analytics if available
      if (window.gtag) {
        window.gtag('consent', 'update', {
          analytics_storage: 'granted',
          ad_storage: 'granted',
        });
      }
    }
  }, []);

  const rejectNonEssential = useCallback(() => {
    const newPreferences = {
      essential: true,
      analytics: false,
      marketing: false,
    };
    
    setConsent('rejected');
    setPreferences(newPreferences);
    
    localStorage.setItem('cookies-consent', 'rejected');
    localStorage.setItem('cookies-preferences', JSON.stringify(newPreferences));
    
    // Disable analytics and marketing cookies
    if (typeof window !== 'undefined') {
      if (window.gtag) {
        window.gtag('consent', 'update', {
          analytics_storage: 'denied',
          ad_storage: 'denied',
        });
      }
    }
  }, []);

  const updatePreferences = useCallback((newPreferences: Partial<CookiesPreferences>) => {
    const updatedPreferences = {
      ...preferences,
      ...newPreferences,
      essential: true, // Always true
    };
    
    setConsent('customized');
    setPreferences(updatedPreferences);
    
    localStorage.setItem('cookies-consent', 'customized');
    localStorage.setItem('cookies-preferences', JSON.stringify(updatedPreferences));
    
    // Update consent based on preferences
    if (typeof window !== 'undefined' && window.gtag) {
      window.gtag('consent', 'update', {
        analytics_storage: updatedPreferences.analytics ? 'granted' : 'denied',
        ad_storage: updatedPreferences.marketing ? 'granted' : 'denied',
      });
    }
  }, [preferences]);

  const clearConsent = useCallback(() => {
    setConsent(null);
    setPreferences({
      essential: true,
      analytics: false,
      marketing: false,
    });
    
    localStorage.removeItem('cookies-consent');
    localStorage.removeItem('cookies-preferences');
  }, []);

  const hasConsented = useCallback(() => {
    return consent !== null;
  }, [consent]);

  const isAnalyticsEnabled = useCallback(() => {
    return preferences.analytics;
  }, [preferences.analytics]);

  const isMarketingEnabled = useCallback(() => {
    return preferences.marketing;
  }, [preferences.marketing]);

  return {
    consent,
    preferences,
    isLoaded,
    acceptAll,
    rejectNonEssential,
    updatePreferences,
    clearConsent,
    hasConsented,
    isAnalyticsEnabled,
    isMarketingEnabled,
  };
};

// Extend Window interface for gtag
declare global {
  interface Window {
    gtag?: (...args: any[]) => void;
  }
}

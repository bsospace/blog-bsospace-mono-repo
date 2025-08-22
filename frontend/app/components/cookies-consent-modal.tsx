'use client'

import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { X, Cookie, Shield, Settings, ChevronDown, ChevronUp, FileText } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';
import { useCookiesConsent } from '@/hooks/use-cookies-consent';
import { CookiesPolicyDialog } from './cookies-policy-dialog';

export const CookiesConsentModal: React.FC = () => {
  const [isVisible, setIsVisible] = useState(false);
  const [isExpanded, setIsExpanded] = useState(false);
  const [showCustomize, setShowCustomize] = useState(false);
  
  const {
    consent,
    preferences,
    isLoaded,
    acceptAll,
    rejectNonEssential,
    updatePreferences,
    hasConsented
  } = useCookiesConsent();

  useEffect(() => {
    // Only show modal if user hasn't consented yet
    if (isLoaded && !hasConsented()) {
      const timer = setTimeout(() => {
        setIsVisible(true);
      }, 1000);
      return () => clearTimeout(timer);
    }
  }, [isLoaded, hasConsented]);

  const handleAccept = () => {
    acceptAll();
    setIsVisible(false);
  };

  const handleReject = () => {
    rejectNonEssential();
    setIsVisible(false);
  };

  const handleCustomize = () => {
    if (showCustomize) {
      // Save current preferences
      updatePreferences(preferences);
      setIsVisible(false);
    } else {
      setShowCustomize(true);
    }
  };

  const handleClose = () => {
    setIsVisible(false);
  };

  const handlePreferenceChange = (key: keyof typeof preferences, value: boolean) => {
    if (key === 'essential') return; // Essential cookies cannot be disabled
    
    updatePreferences({
      [key]: value
    });
  };

  if (!isVisible || !isLoaded) return null;

  return (
    <div className="fixed bottom-0 left-0 right-0 z-50 p-4 flex items-center justify-end drop-shadow-lg">
      <Card className="max-w-4xl mx-auto shadow-2xl">
        <CardContent className="p-6">
          {/* Header */}
          <div className="flex items-start justify-between mb-6">
            <div className="flex items-center gap-4">
              <div className="p-3 bg-primary rounded-2xl shadow-lg">
                <Cookie className="h-6 w-6 text-primary-foreground" />
              </div>
              <div>
                <h3 className="text-xl font-bold text-foreground mb-1">
                  Cookie Preferences
                </h3>
                <p className="text-sm text-muted-foreground">
                  We use cookies to enhance your browsing experience
                </p>
              </div>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleClose}
              className="rounded-full p-2"
            >
              <X className="h-5 w-5" />
            </Button>
          </div>

          {/* Main Content */}
          <div className="space-y-6">
            <p className="text-foreground leading-relaxed text-base">
              This website uses cookies to ensure you get the best experience. We use essential cookies for basic functionality, 
              analytics cookies to understand how you use our site, and marketing cookies to provide personalized content.
            </p>

            {/* Cookie Categories - Always visible when customizing */}
            {(isExpanded || showCustomize) && (
              <div className="space-y-4 pt-4 border-t border-border">
                {/* Essential Cookies */}
                <div className="flex items-center justify-between p-4 bg-muted/50 rounded-xl border border-border">
                  <div className="flex items-center gap-4">
                    <div className="p-2 bg-primary/10 rounded-lg">
                      <Shield className="h-5 w-5 text-primary" />
                    </div>
                    <div className="flex-1">
                      <p className="font-semibold text-foreground">Essential Cookies</p>
                      <p className="text-sm text-muted-foreground">Required for basic site functionality</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <Badge variant="secondary">
                      Always Active
                    </Badge>
                    {showCustomize && (
                      <Checkbox
                        checked={preferences.essential}
                        disabled
                        className="ml-2"
                      />
                    )}
                  </div>
                </div>

                {/* Analytics Cookies */}
                <div className="flex items-center justify-between p-4 bg-muted/50 rounded-xl border border-border">
                  <div className="flex items-center gap-4">
                    <div className="p-2 bg-secondary/20 rounded-lg">
                      <Settings className="h-5 w-5 text-secondary-foreground" />
                    </div>
                    <div className="flex-1">
                      <p className="font-semibold text-foreground">Analytics Cookies</p>
                      <p className="text-sm text-muted-foreground">Help us improve our website</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <Badge variant="outline">
                      Optional
                    </Badge>
                    {showCustomize && (
                      <Checkbox
                        checked={preferences.analytics}
                        onCheckedChange={(checked) => 
                          handlePreferenceChange('analytics', checked as boolean)
                        }
                        className="ml-2"
                      />
                    )}
                  </div>
                </div>

                {/* Marketing Cookies */}
                <div className="flex items-center justify-between p-4 bg-muted/50 rounded-xl border border-border">
                  <div className="flex items-center gap-4">
                    <div className="p-2 bg-accent/20 rounded-lg">
                      <Cookie className="h-5 w-5 text-accent-foreground" />
                    </div>
                    <div className="flex-1">
                      <p className="font-semibold text-foreground">Marketing Cookies</p>
                      <p className="text-sm text-muted-foreground">Personalized content and ads</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <Badge variant="outline">
                      Optional
                    </Badge>
                    {showCustomize && (
                      <Checkbox
                        checked={preferences.marketing}
                        onCheckedChange={(checked) => 
                          handlePreferenceChange('marketing', checked as boolean)
                        }
                        className="ml-2"
                      />
                    )}
                  </div>
                </div>
              </div>
            )}

            {/* Expandable Details */}
            {!showCustomize && (
              <div className="pt-2">
                <Button
                  variant="ghost"
                  onClick={() => setIsExpanded(!isExpanded)}
                  className="text-primary hover:text-primary/80 hover:bg-primary/10 p-0 h-auto font-medium"
                >
                  {isExpanded ? (
                    <>
                      <ChevronUp className="h-4 w-4 mr-2" />
                      Hide Details
                    </>
                  ) : (
                    <>
                      <ChevronDown className="h-4 w-4 mr-2" />
                      Show Details
                    </>
                  )}
                </Button>
              </div>
            )}

            {/* Action Buttons */}
            <div className="flex flex-col sm:flex-row gap-4 pt-6">
              {!showCustomize ? (
                <>
                  <Button
                    onClick={handleAccept}
                    className="flex-1 py-3 rounded-xl shadow-lg hover:shadow-xl transition-all duration-200 transform hover:scale-105"
                  >
                    Accept All Cookies
                  </Button>
                  
                  <Button
                    onClick={handleCustomize}
                    variant="outline"
                    className="flex-1 py-3 rounded-xl transition-all duration-200"
                  >
                    Customize Settings
                  </Button>
                  
                  <Button
                    onClick={handleReject}
                    variant="destructive"
                    className="flex-1 py-3 rounded-xl transition-all duration-200"
                  >
                    Reject Non-Essential
                  </Button>
                </>
              ) : (
                <>
                  <Button
                    onClick={handleCustomize}
                    className="flex-1 py-3 rounded-xl shadow-lg hover:shadow-xl transition-all duration-200 transform hover:scale-105"
                  >
                    Save Preferences
                  </Button>
                  
                  <Button
                    onClick={() => setShowCustomize(false)}
                    variant="outline"
                    className="flex-1 py-3 rounded-xl transition-all duration-200"
                  >
                    Back
                  </Button>
                </>
              )}
            </div>

            {/* Footer */}
            <div className="pt-6 border-t border-border">
              <div className="flex flex-col sm:flex-row items-center justify-between gap-4 mb-4">
                <p className="text-xs text-muted-foreground text-center sm:text-left leading-relaxed">
                  By continuing to use this site, you consent to our use of cookies. 
                  Learn more in our{' '}
                  <a href="/privacy" className="text-primary hover:underline font-medium">
                    Privacy Policy
                  </a>
                  {' '}and{' '}
                  <CookiesPolicyDialog>
                    <button className="text-primary hover:underline font-medium cursor-pointer">
                      Cookie Policy
                    </button>
                  </CookiesPolicyDialog>
                  .
                </p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default CookiesConsentModal;

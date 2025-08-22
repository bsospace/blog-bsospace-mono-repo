'use client'

import React, { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';
import { Separator } from '@/components/ui/separator';
import { Cookie, Shield, Settings, RefreshCw } from 'lucide-react';
import { useCookiesConsentContext } from '@/app/contexts/cookies-consent-context';

export const CookiesPreferences: React.FC = () => {
  const [isUpdating, setIsUpdating] = useState(false);
  const {
    consent,
    preferences,
    updatePreferences,
    clearConsent,
    hasConsented
  } = useCookiesConsentContext();

  const handlePreferenceChange = async (key: keyof typeof preferences, value: boolean) => {
    if (key === 'essential') return; // Essential cookies cannot be disabled
    
    setIsUpdating(true);
    try {
      updatePreferences({
        [key]: value
      });
    } finally {
      setIsUpdating(false);
    }
  };

  const handleClearConsent = () => {
    clearConsent();
  };

  if (!hasConsented()) {
    return (
      <Card className="w-full">
        <CardHeader>
          <CardTitle className="flex items-center gap-3">
            <div className="p-2 bg-primary rounded-xl">
              <Cookie className="h-5 w-5 text-primary-foreground" />
            </div>
            Cookie Preferences
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground">
            You haven't set your cookie preferences yet. Please accept or reject cookies to continue.
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle className="flex items-center gap-3">
          <div className="p-2 bg-primary rounded-xl">
            <Cookie className="h-5 w-5 text-primary-foreground" />
          </div>
          Cookie Preferences
        </CardTitle>
        <div className="flex items-center gap-3">
          <Badge 
            variant={consent === 'accepted' ? 'default' : 'secondary'}
          >
            {consent === 'accepted' ? 'All Accepted' : 
             consent === 'rejected' ? 'Minimal' : 'Customized'}
          </Badge>
          {isUpdating && <RefreshCw className="h-4 w-4 animate-spin text-primary" />}
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Essential Cookies */}
        <div className="flex items-center justify-between p-4 bg-muted/50 rounded-xl border border-border">
          <div className="flex items-center gap-4">
            <div className="p-2 bg-primary/10 rounded-lg">
              <Shield className="h-5 w-5 text-primary" />
            </div>
            <div>
              <p className="font-semibold text-foreground">Essential Cookies</p>
              <p className="text-sm text-muted-foreground">Required for basic site functionality</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <Badge variant="secondary">
              Always Active
            </Badge>
            <Checkbox
              checked={preferences.essential}
              disabled
              className="ml-2"
            />
          </div>
        </div>

        <Separator />

        {/* Analytics Cookies */}
        <div className="flex items-center justify-between p-4 bg-muted/50 rounded-xl border border-border">
          <div className="flex items-center gap-4">
            <div className="p-2 bg-secondary/20 rounded-lg">
              <Settings className="h-5 w-5 text-secondary-foreground" />
            </div>
            <div>
              <p className="font-semibold text-foreground">Analytics Cookies</p>
              <p className="text-sm text-muted-foreground">Help us improve our website</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <Badge 
              variant={preferences.analytics ? 'default' : 'outline'}
            >
              {preferences.analytics ? 'Enabled' : 'Disabled'}
            </Badge>
            <Checkbox
              checked={preferences.analytics}
              onCheckedChange={(checked) => 
                handlePreferenceChange('analytics', checked as boolean)
              }
              disabled={isUpdating}
              className="ml-2"
            />
          </div>
        </div>

        <Separator />

        {/* Marketing Cookies */}
        <div className="flex items-center justify-between p-4 bg-muted/50 rounded-xl border border-border">
          <div className="flex items-center gap-4">
            <div className="p-2 bg-accent/20 rounded-lg">
              <Cookie className="h-5 w-5 text-accent-foreground" />
            </div>
            <div>
              <p className="font-semibold text-foreground">Marketing Cookies</p>
              <p className="text-sm text-muted-foreground">Personalized content and ads</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <Badge 
              variant={preferences.marketing ? 'default' : 'outline'}
            >
              {preferences.marketing ? 'Enabled' : 'Disabled'}
            </Badge>
            <Checkbox
              checked={preferences.marketing}
              onCheckedChange={(checked) => 
                handlePreferenceChange('marketing', checked as boolean)
              }
              disabled={isUpdating}
              className="ml-2"
            />
          </div>
        </div>

        <Separator />

        {/* Actions */}
        <div className="flex justify-between items-center pt-4">
          <p className="text-sm text-muted-foreground">
            Last updated: {new Date().toLocaleDateString()}
          </p>
          <Button
            variant="destructive"
            onClick={handleClearConsent}
            disabled={isUpdating}
            className="font-medium rounded-xl transition-all duration-200"
          >
            Reset Preferences
          </Button>
        </div>
      </CardContent>
    </Card>
  );
};

export default CookiesPreferences;

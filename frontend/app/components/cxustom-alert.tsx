import React, { useEffect, useState } from 'react';
import { AlertCircle, CheckCircle, Info, X, XCircle } from 'lucide-react';
import { cn } from '@/lib/utils';

export type AlertType = 'success' | 'error' | 'warning' | 'info';

export interface AlertProps {
  type: AlertType;
  title: string;
  message?: string;
  duration?: number; // Auto-hide duration in milliseconds
  onClose?: () => void;
  show?: boolean;
}

const alertStyles = {
  success: {
    container: 'bg-green-50 border-green-200 text-green-800 dark:bg-green-900/20 dark:border-green-800 dark:text-green-300',
    icon: 'text-green-500 dark:text-green-400',
    closeButton: 'hover:bg-green-100 dark:hover:bg-green-800/30'
  },
  error: {
    container: 'bg-red-50 border-red-200 text-red-800 dark:bg-red-900/20 dark:border-red-800 dark:text-red-300',
    icon: 'text-red-500 dark:text-red-400',
    closeButton: 'hover:bg-red-100 dark:hover:bg-red-800/30'
  },
  warning: {
    container: 'bg-yellow-50 border-yellow-200 text-yellow-800 dark:bg-yellow-900/20 dark:border-yellow-800 dark:text-yellow-300',
    icon: 'text-yellow-500 dark:text-yellow-400',
    closeButton: 'hover:bg-yellow-100 dark:hover:bg-yellow-800/30'
  },
  info: {
    container: 'bg-blue-50 border-blue-200 text-blue-800 dark:bg-blue-900/20 dark:border-blue-800 dark:text-blue-300',
    icon: 'text-blue-500 dark:text-blue-400',
    closeButton: 'hover:bg-blue-100 dark:hover:bg-blue-800/30'
  }
};

const alertIcons = {
  success: CheckCircle,
  error: XCircle,
  warning: AlertCircle,
  info: Info
};

export const CustomAlert: React.FC<AlertProps> = ({
  type,
  title,
  message,
  duration = 5000,
  onClose,
  show = true
}) => {
  const [isVisible, setIsVisible] = useState(show);
  const [isAnimating, setIsAnimating] = useState(false);

  const styles = alertStyles[type];
  const Icon = alertIcons[type];

  useEffect(() => {
    setIsVisible(show);
    if (show) {
      setIsAnimating(true);
    }
  }, [show]);

  useEffect(() => {
    if (duration > 0 && isVisible) {
      const timer = setTimeout(() => {
        handleClose();
      }, duration);

      return () => clearTimeout(timer);
    }
  }, [duration, isVisible]);

  const handleClose = () => {
    setIsAnimating(false);
    setTimeout(() => {
      setIsVisible(false);
      onClose?.();
    }, 200);
  };

  if (!isVisible) return null;

  return (
    <div className="fixed top-4 right-4 z-50 max-w-sm w-full">
      <div
        className={cn(
          'border rounded-lg p-4 shadow-lg transition-all duration-200 ease-in-out',
          styles.container,
          isAnimating ? 'translate-x-0 opacity-100' : 'translate-x-full opacity-0'
        )}
      >
        <div className="flex items-start gap-3">
          <Icon className={cn('h-5 w-5 mt-0.5 flex-shrink-0', styles.icon)} />
          
          <div className="flex-1 min-w-0">
            <h4 className="font-medium text-sm leading-5">{title}</h4>
            {message && (
              <p className="mt-1 text-sm leading-5 opacity-90">{message}</p>
            )}
          </div>

          <button
            onClick={handleClose}
            className={cn(
              'p-1 rounded-md transition-colors duration-200 flex-shrink-0',
              styles.closeButton
            )}
            aria-label="Close alert"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        {/* Progress bar for auto-hide */}
        {duration > 0 && (
          <div className="mt-3 w-full bg-current opacity-20 rounded-full h-1">
            <div
              className="bg-current h-1 rounded-full transition-all duration-200 ease-linear"
              style={{
                width: isAnimating ? '100%' : '0%',
                transitionDuration: `${duration}ms`
              }}
            />
          </div>
        )}
      </div>
    </div>
  );
};

// Alert Manager for multiple alerts
export const AlertManager: React.FC = () => {
  const [alerts, setAlerts] = useState<Array<AlertProps & { id: string }>>([]);

  const addAlert = (alert: Omit<AlertProps, 'show'>) => {
    const id = Math.random().toString(36).substr(2, 9);
    const newAlert = { ...alert, id, show: true };
    setAlerts(prev => [...prev, newAlert]);
    return id;
  };

  const removeAlert = (id: string) => {
    setAlerts(prev => prev.filter(alert => alert.id !== id));
  };

  const updateAlert = (id: string, updates: Partial<AlertProps>) => {
    setAlerts(prev => prev.map(alert => 
      alert.id === id ? { ...alert, ...updates } : alert
    ));
  };

  return (
    <div className="fixed top-4 right-4 z-50 space-y-3">
      {alerts.map((alert) => (
        <CustomAlert
          key={alert.id}
          {...alert}
          onClose={() => removeAlert(alert.id)}
        />
      ))}
    </div>
  );
};

// Hook for easy alert usage
export const useAlert = () => {
  const [alerts, setAlerts] = useState<Array<AlertProps & { id: string }>>([]);

  const showAlert = (alert: Omit<AlertProps, 'show'>) => {
    const id = Math.random().toString(36).substr(2, 9);
    const newAlert = { ...alert, id, show: true };
    setAlerts(prev => [...prev, newAlert]);
    return id;
  };

  const hideAlert = (id: string) => {
    setAlerts(prev => prev.map(alert => 
      alert.id === id ? { ...alert, show: false } : alert
    ));
  };

  const removeAlert = (id: string) => {
    setTimeout(() => {
      setAlerts(prev => prev.filter(alert => alert.id !== id));
    }, 200);
  };

  const success = (title: string, message?: string, duration?: number) => {
    const id = showAlert({ type: 'success', title, message, duration });
    setTimeout(() => hideAlert(id), duration || 5000);
    setTimeout(() => removeAlert(id), (duration || 5000) + 200);
  };

  const error = (title: string, message?: string, duration?: number) => {
    const id = showAlert({ type: 'error', title, message, duration });
    setTimeout(() => hideAlert(id), duration || 5000);
    setTimeout(() => removeAlert(id), (duration || 5000) + 200);
  };

  const warning = (title: string, message?: string, duration?: number) => {
    const id = showAlert({ type: 'warning', title, message, duration });
    setTimeout(() => hideAlert(id), duration || 5000);
    setTimeout(() => removeAlert(id), (duration || 5000) + 200);
  };

  const info = (title: string, message?: string, duration?: number) => {
    const id = showAlert({ type: 'info', title, message, duration });
    setTimeout(() => hideAlert(id), duration || 5000);
    setTimeout(() => removeAlert(id), (duration || 5000) + 200);
  };

  return {
    alerts,
    success,
    error,
    warning,
    info,
    showAlert,
    hideAlert,
    removeAlert
  };
};

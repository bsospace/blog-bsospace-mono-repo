interface IEvnConfig {
    [key: string]: string | number | undefined;
    nodeEnv: string;
    apiBaseUrl: string;
    openIdApiUrl: string;
    callBackUrl: string;
    ws: string;
    domain: string;
    email: string;
    organizationName: string;
    imageServiceUrl: string;
    proxyUrl: string;
    contactPersonName: string;
}

const envConfig: IEvnConfig = {
    nodeEnv: process.env.NODE_ENV || 'development',
    apiBaseUrl: process.env.NEXT_PUBLIC_API_BASE_URL || '',
    openIdApiUrl: process.env.NEXT_PUBLIC_OPENID_API_URL || '',
    callBackUrl: process.env.NEXT_PUBLIC_CALLBACK_URL || '',
    ws: process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8088/api/v1/ws',
    domain: process.env.NEXT_PUBLIC_DOMAIN || '',
    email: process.env.NEXT_PUBLIC_EMAIL || '',
    organizationName: process.env.NEXT_PUBLIC_ORGANIZATION_NAME || 'BSO Space',
    imageServiceUrl: process.env.NEXT_PUBLIC_IMAGE_SERVICE_URL || '',
    proxyUrl: process.env.NEXT_PUBLIC_PROXY_URL || '',
    contactPersonName: process.env.NEXT_PUBLIC_CONTACT_PERSON_NAME || ''
};


export const checkEnvConfig = () => {
    // Define only the required fields
    const requiredFields: string[] = [
        'nodeEnv',
        'apiBaseUrl',
        'openIdApiUrl',
        'callBackUrl',
        'ws',
        'domain',
        'email',
        'organizationName',
        'imageServiceUrl',
        'proxyUrl',
        'contactPersonName'
    ];

    for (const field of requiredFields) {
        if (!envConfig[field] || envConfig[field] === '') {
            throw new Error(`Missing required environment variable: ${field.toUpperCase()}`);
        }
    }
};

try {
    checkEnvConfig();
} catch (error) {
    if (error instanceof Error) {
        console.error(error.message)
    } else {
        console.error('An unknown error occurred.');
    }
}

export default envConfig;
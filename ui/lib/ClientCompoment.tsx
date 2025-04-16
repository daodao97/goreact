import React, { Suspense, useState, useEffect } from 'react';

interface ClientCompomentProps {
    fallback: React.ReactNode;
    children: React.ReactNode;
}

export const ClientCompoment: React.FC<ClientCompomentProps> = ({ fallback, children }) => {
    const [isClient, setIsClient] = useState(false);

    useEffect(() => {
        setIsClient(true);
    }, []);

    return (
        <Suspense fallback={fallback}>
            {isClient ? children : fallback}
        </Suspense>
    );
};

export default ClientCompoment;

import React from 'react';

interface FeatureLayoutProps {
    title: string;
    description: string;
    children: React.ReactNode;
}

export const FeatureLayout: React.FC<FeatureLayoutProps> = ({ title, description, children }) => {
    return (
        <div className="py-16">
            <div className="container mx-auto px-4">
                {title && <h2 className="text-3xl font-bold mb-4">{title}</h2>}
                {description && <p className="text-gray-600 mb-8">{description}</p>}
                {children}
            </div>
        </div>
    );
}

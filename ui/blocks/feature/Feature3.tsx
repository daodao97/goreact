import React from 'react';

interface Feature3Props {
    title: string;
    description: string;
    features: {
        title: string;
        description: string;
    }[];
    alignment?: 'left' | 'center' | 'right';
}

export default function Feature3({ title, description, features, alignment = 'center' }: Feature3Props) {
    const getAlignmentClass = () => {
        switch (alignment) {
            case 'left':
                return '';
            case 'center':
                return 'mx-auto';
            case 'right':
                return 'ml-auto';
            default:
                return 'mx-auto';
        }
    };

    const getTextAlignmentClass = () => {
        switch (alignment) {
            case 'left':
                return 'text-left';
            case 'center':
                return 'text-center';
            case 'right':
                return 'text-right';
            default:
                return 'text-center';
        }
    };

    return (
        <div className={`max-w-7xl ${getAlignmentClass()} px-4 sm:px-6 lg:px-8 pt-12`}>
            <h2 className={`text-3xl font-bold mb-4 text-gray-900 dark:text-white ${getTextAlignmentClass()}`}>{title}</h2>
            {description && <p className={`text-gray-600 dark:text-gray-300 mb-8 ${getTextAlignmentClass()}`}>{description}</p>}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 md:gap-8">
                {features.map((feature, index) => (
                    <div key={index} className="p-6 bg-white dark:bg-gray-800 rounded-lg shadow border border-gray-200 dark:border-gray-700">
                        <h3 className="text-xl font-semibold mb-2 text-gray-900 dark:text-white">{feature.title}</h3>
                        <p className="text-gray-600 dark:text-gray-300">{feature.description}</p>
                    </div>
                ))}
            </div>
        </div>
    );
}

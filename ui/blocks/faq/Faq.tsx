import React, { useState } from "react";
import { Badge } from "@radix-ui/themes";

type Faq = {
    tips: string;
    title: string;
    subtitle: string;
    list: FaqItem[];
    anchor: string;
};

type FaqItem = {
    question: string;
    answer: string;
};

export default function Faq({ faqs }: { faqs: Faq }) {
    return (
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-12">
            <div className="text-center mb-12" id={faqs.anchor}>
                {faqs.tips && <span className="inline-block bg-amber-500 px-3 py-1 rounded-full text-sm font-medium mb-4">{faqs.tips}</span>}
                <h2 className="text-3xl font-bold sm:text-4xl mb-4 text-gray-900 dark:text-white">{faqs.title}</h2>
                {faqs.subtitle && <p className="max-w-2xl mx-auto text-gray-600 dark:text-gray-300">{faqs.subtitle}</p>}
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 lg:gap-8">
                {(faqs.list || []).map((faq, index) => (
                    <div key={index} className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6 relative border border-gray-200 dark:border-gray-700">
                        <div className="flex flex-col items-start">
                            <div className="flex flex-row items-center">
                                <Badge color="orange" className="mr-2 flex items-center">{index + 1}</Badge>
                                <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-0">{faq.question}</h3>
                            </div>
                            <p className="text-gray-600 dark:text-gray-400">{faq.answer}</p>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}

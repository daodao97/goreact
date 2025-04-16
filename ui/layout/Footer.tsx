import React from "react";
import { getWebsite, t, getTranslations } from "../lib/i18n";

import { FaGithub, FaTwitter, FaFacebook, FaInstagram, FaLinkedin, FaYoutube } from "react-icons/fa";

const socialIcons = {
    twitter: FaTwitter,
    facebook: FaFacebook,
    instagram: FaInstagram,
    linkedin: FaLinkedin,
    youtube: FaYoutube,
    github: FaGithub,
}

export function Footer() {
    const website = getTranslations("root.footer", {
        logo: "",
        title: "",
        desc: "",
        social: [],
        links: [],
        copyright: "",
        policy: [],
    })
    return (
        <footer className="py-12 w-full bg-gray-50 dark:bg-gray-900">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div className="flex flex-col md:flex-row justify-between md:items-center">
                    {/* Logo and Description Section */}
                    <div className="flex flex-col items-center md:items-start text-center md:text-left">
                        <div className="flex items-center">
                            {website.logo && (
                                <img src={website.logo} alt={t(website.title, website.title)} className="h-8 w-8" />
                            )}
                            <span className={website.logo ? 'ml-2' : '' + ' text-xl text-amber-600 font-semibold'}>{t(website.title, website.title)}</span>
                        </div>
                        <p className="mt-4 text-gray-600 dark:text-gray-400">
                            {t(website.desc, website.desc)}
                        </p>
                        <div className="flex space-x-4 mt-4">
                            {(website.social || []).map((social) => (
                                <a target="_blank" key={social.icon + " " + social.url} href={social.url} className="text-gray-600">
                                    {social.icon ? React.createElement(socialIcons[social.icon as keyof typeof socialIcons]) : social.text}
                                </a>
                            ))}
                        </div>
                    </div>

                    {/* Navigation Links */}
                    <div className="mt-8 md:mt-0 md:col-span-6 md:col-start-7 text-center md:text-right">
                        <div className="flex flex-row justify-center md:justify-end overflow-x-auto md:flex md:flex-row md:gap-8">
                            {(website.links || []).map((group) => (
                                <div key={group.title} className="text-center md:text-right min-w-max px-3">
                                    <h2 className="text-white font-medium">{t(group.title, group.title)}</h2>
                                    <ul className="mt-4 space-y-2">
                                        {(group.links || []).map((link) => (
                                            <li key={link.text}>
                                                <a
                                                    href={link.url.startsWith('#') ? link.url : link.url}
                                                    target={link.url.startsWith('#') ? '_self' : '_blank'}
                                                    className="text-gray-600"
                                                >
                                                    {link.text}
                                                </a>
                                            </li>
                                        ))}
                                    </ul>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>

                {/* Copyright Section */}
                <div className="mt-8 pt-8 border-t pb-4 border-gray-700 dark:border-gray-900">
                    <div className="flex flex-col md:flex-row justify-between items-center">
                        <p className="text-gray-500 text-sm">{t(website.copyright, website.copyright)}</p>
                        <div className="mt-4 md:mt-0">
                            {(website.policy || []).map((policy, index) => (
                                <React.Fragment key={policy.url}>
                                    <a href={policy.url}
                                        className="text-gray-500 text-sm">{t(policy.text, policy.text)}</a>
                                    {index < website.policy.length - 1 && <span className="mx-2 text-gray-500">|</span>}
                                </React.Fragment>
                            ))}
                        </div>
                    </div>
                </div>
            </div>
        </footer >
    );
}

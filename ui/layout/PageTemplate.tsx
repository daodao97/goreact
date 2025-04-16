import React, { ReactNode } from "react";
import { Box, Heading, Text } from "@radix-ui/themes";

interface PageTemplateProps {
    title: string;
    description?: string;
    children: ReactNode;
    headerActions?: ReactNode;
    footerContent?: ReactNode;
}

/**
 * 统一的页面模板组件
 * 
 * 提供一致的页面布局结构，包括标题、描述、内容区域、页头操作和页脚内容
 */
export function PageTemplate({
    title,
    description,
    children,
    headerActions,
    footerContent
}: PageTemplateProps) {
    return (
        <Box className="page-container py-4">
            <Box className="page-header mb-6">
                <Box className="flex justify-between items-center mb-2">
                    <Heading size="6" className="page-title">{title}</Heading>
                    {headerActions && (
                        <Box className="page-header-actions">
                            {headerActions}
                        </Box>
                    )}
                </Box>
                {description && (
                    <Text className="page-description text-gray-500">
                        {description}
                    </Text>
                )}
            </Box>

            <Box className="page-content mb-8">
                {children}
            </Box>

            {footerContent && (
                <Box className="page-footer mt-auto pt-4 border-t border-gray-200">
                    {footerContent}
                </Box>
            )}
        </Box>
    );
} 
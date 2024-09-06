export interface PluginRegistry {
    registerRootComponent(DMRestrictionsComponent: () => JSX.Element): unknown;
    registerPostTypeComponent(typeName: string, component: React.ElementType)

    // Add more if needed from https://developers.mattermost.com/extend/plugins/webapp/reference
}

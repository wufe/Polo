import React from 'react';
import { IApp } from '@/state/models/app-model';
import { useParams } from 'react-router';
import { Content } from '@/components/shared/ui-elements/layout/content';

type TProps = {
    app: IApp;
}
export const ApplicationEditPage = (props: TProps) => {

    const { id } = useParams<{ id: string; }>();

    return <>
        <Content>
            <h1>Edit application</h1>
            <span>{id}</span>
        </Content>
    </>
};

export default ApplicationEditPage;
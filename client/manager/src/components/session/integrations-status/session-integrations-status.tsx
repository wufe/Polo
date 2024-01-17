import { ISession } from '@polo/common/state/models';
import { observer } from 'mobx-react-lite';
import React from 'react';
import { TiltSessionIntegrationStatusDashboardLink } from './tilt/tilt-session-integration-status-dashboard-link';

type TProps = {
    session: ISession;
    integrationsStatus: ISession['integrations'];
};

export const SessionIntegrationsStatus = observer((props: TProps) => {
    return <TiltSessionIntegrationStatusDashboardLink
        session={props.session}
        dashboards={props.integrationsStatus.tilt.dashboards} />;
});
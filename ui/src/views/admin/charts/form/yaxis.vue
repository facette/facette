<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<!-- eslint-disable vue/no-mutating-props -->

<template>
    <v-checkbox type="toggle" v-model:value="axis.show">
        {{ i18n.t("labels.show._") }}
    </v-checkbox>

    <div class="reveal" :class="{visible: axis.show}">
        <v-label>{{ i18n.t("labels.labels._", 1) }}</v-label>
        <v-input :delay="350" :help="i18n.t('help.charts.axes.label')" v-model:value="axis.label"></v-input>

        <v-flex>
            <v-flex direction="column">
                <v-label>{{ i18n.t("labels.charts.axes.min") }}</v-label>
                <v-input
                    type="number"
                    :delay="350"
                    :help="i18n.t('help.charts.axes.min')"
                    v-model:value.number="axis.min"
                ></v-input>
            </v-flex>

            <v-flex direction="column">
                <v-label>{{ i18n.t("labels.charts.axes.max") }}</v-label>
                <v-input
                    type="number"
                    :delay="350"
                    :help="i18n.t('help.charts.axes.max')"
                    v-model:value.number="axis.max"
                ></v-input>
            </v-flex>
        </v-flex>
    </div>
</template>

<script lang="ts">
import {useI18n} from "vue-i18n";

import {ChartYAxis} from "types/api";

export default {
    inheritAttrs: false,
    props: {
        axis: {
            required: true,
            type: Object as () => ChartYAxis,
        },
    },
    setup(): Record<string, unknown> {
        const i18n = useI18n();

        return {
            i18n,
        };
    },
};
</script>

<style lang="scss" scoped>
@import "../../../../components/ui/components/form/mixins";

.reveal {
    @include form;

    visibility: hidden;

    &.visible {
        visibility: visible;
    }

    .v-flex.column + .column {
        margin-left: var(--content-padding);
    }
}
</style>

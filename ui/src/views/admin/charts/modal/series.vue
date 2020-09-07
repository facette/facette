<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <v-modal name="chart-series" :title="i18n.t(`labels.series.${edit ? 'edit' : 'add'}`)" @show="onShow">
        <template v-slot="modal">
            <v-form>
                <v-label>{{ i18n.t("labels.series._", 1) }}</v-label>
                <v-editor required v-autofocus v-model:value="series.expr" @update:value="onUpdate"></v-editor>

                <v-toolbar>
                    <v-label v-if="loadingMetrics">
                        <v-spinner
                            :size="16"
                            :stroke-width="2"
                            :style="{'--accent': 'var(--color)', '--spinner-background': 'var(--input-background)'}"
                        ></v-spinner>

                        <span>{{ i18n.t("labels.metrics.fetching") }}</span>
                    </v-label>

                    <template v-else-if="error">
                        <v-label icon="exclamation-triangle">{{ error }}</v-label>
                    </template>

                    <template v-else>
                        <v-label icon="check">{{ i18n.t("labels.ok") }}</v-label>

                        <v-spacer></v-spacer>

                        <v-label icon="dot-circle" v-if="metricsCount !== null">
                            {{ i18n.t("labels.metrics.matching", [metricsCount], metricsCount) }}
                        </v-label>
                    </template>
                </v-toolbar>

                <v-flex class="columns">
                    <v-flex direction="column">
                        <v-label>{{ i18n.t("labels.color") }}</v-label>
                        <v-color class="half" v-model:value="series.options.color"></v-color>
                    </v-flex>

                    <v-flex direction="column">
                        <v-label>{{ i18n.t("labels.charts.axes._", 1) }}</v-label>
                        <v-select
                            class="half"
                            required
                            :options="axes"
                            :placeholder="i18n.t('labels.charts.axes.select')"
                            v-model:value="series.options.axis"
                        ></v-select>
                    </v-flex>
                </v-flex>

                <template v-slot:bottom>
                    <v-button icon="indent" @click="format" @mousedown.prevent>
                        {{ i18n.t("labels.format") }}
                    </v-button>

                    <v-spacer></v-spacer>

                    <v-button @click="modal.close(false)">
                        {{ i18n.t("labels.cancel") }}
                    </v-button>

                    <v-button
                        :disabled="!series.expr || !series.options.axis || Boolean(error)"
                        primary
                        @click="modal.close(series)"
                    >
                        {{ i18n.t("labels.series.set") }}
                    </v-button>
                </template>
            </v-form>
        </template>
    </v-modal>
</template>

<script lang="ts">
import cloneDeep from "lodash/cloneDeep";
import merge from "lodash/merge";
import {ref} from "vue";
import {useI18n} from "vue-i18n";

import {SelectOption} from "types/ui";

import api from "@/lib/api";
import {formatExpr, parseExpr} from "@/lib/expr";

export interface ModalChartSeriesParams {
    edit: boolean;
    series: ChartSeries;
}

const defaultSeries: ChartSeries = {
    expr: "",
    options: {
        color: "",
        axis: "left",
    },
};

export default {
    setup(): Record<string, unknown> {
        const i18n = useI18n();

        let checkTimeout: number | null = null;

        const edit = ref(false);
        const error = ref<string | null>(null);
        const loadingMetrics = ref(false);
        const metricsCount = ref<number | null>(null);
        const series = ref<ChartSeries>(cloneDeep(defaultSeries));

        const axes = ref<Array<SelectOption>>([
            {label: i18n.t("labels.charts.axes.left"), value: "left"},
            {label: i18n.t("labels.charts.axes.right"), value: "right"},
        ]);

        const checkExpr = async (value: string): Promise<void> => {
            try {
                const expr = parseExpr(value);

                error.value = null;

                if (expr.type === "matcher") {
                    loadingMetrics.value = true;

                    metricsCount.value = await api
                        .metrics({match: value})
                        .then(
                            response => Promise.resolve(response.total ?? null),
                            () => Promise.resolve(null),
                        )
                        .finally(() => {
                            loadingMetrics.value = false;
                        });
                } else {
                    metricsCount.value = null;
                }
            } catch (e) {
                error.value = e.message;
            }
        };

        const format = (): void => {
            series.value.expr = formatExpr(series.value.expr);
        };

        const onShow = (params: ModalChartSeriesParams): void => {
            edit.value = params.edit;
            series.value = merge({}, defaultSeries, params.series);

            if (series.value.expr) {
                checkExpr(series.value.expr);
            }
        };

        const onUpdate = (value: string): void => {
            if (checkTimeout !== null) {
                clearTimeout(checkTimeout);
            }

            checkTimeout = setTimeout(() => checkExpr(value), 1000);
        };

        return {
            axes,
            edit,
            error,
            format,
            i18n,
            loadingMetrics,
            metricsCount,
            onShow,
            onUpdate,
            series,
        };
    },
};
</script>

<style lang="scss" scoped>
@import "../../../mixins";
@import "../../../../components/ui/components/form/mixins";

.v-modal {
    ::v-deep(.v-modal-content) {
        @include content;

        width: 35vw;

        .v-editor {
            border-radius: 0.2rem 0.2rem 0 0;

            .v-editor-input {
                height: 10rem;
                min-height: 10rem;
                resize: vertical;
            }

            + .v-toolbar {
                background-color: var(--dark-gray);
                border-radius: 0 0 0.2rem 0.2rem;
                height: 2rem;
                line-height: 2rem;
                margin-top: 0;
                padding: 0 0.25rem;

                .v-label {
                    .v-spinner {
                        margin-right: 0.5rem;
                    }

                    span {
                        opacity: 0.5;
                    }
                }

                .v-icon {
                    font-size: 0.8rem;
                }
            }
        }

        .columns .column {
            @include form;
        }
    }
}
</style>

<?php
/**
 * Plugin Name: findable
 * Description: Scan any post or page for SEO, AEO and GEO issues with the free findable.page API.
 * Version: 0.1.0
 * Requires at least: 5.8
 * Requires PHP: 7.4
 * License: GPLv2 or later
 * License URI: https://www.gnu.org/licenses/gpl-2.0.html
 * Text Domain: findable
 */

if ( ! defined( 'ABSPATH' ) ) {
	exit;
}

define( 'FINDABLE_VERSION', '0.1.0' );
define( 'FINDABLE_API', 'https://findable.page/api/scan' );

/**
 * Scan a post's permalink and store the summary in post meta.
 *
 * @return array|WP_Error
 */
function findable_scan_post( $post_id, $keyphrase = '' ) {
	$post = get_post( $post_id );
	if ( ! $post || ! in_array( $post->post_type, array( 'post', 'page' ), true ) || 'publish' !== $post->post_status || ! is_post_publicly_viewable( $post ) ) {
		return new WP_Error( 'findable_unpublished', __( 'Only published posts can be scanned.', 'findable' ) );
	}
	$url  = get_permalink( $post );
	$body = array( 'url' => $url );
	if ( '' !== $keyphrase ) {
		$body['keyphrase'] = $keyphrase;
	}
	$res = wp_remote_post(
		FINDABLE_API,
		array(
			'timeout'             => 30,
			'limit_response_size' => MB_IN_BYTES,
			'headers' => array( 'Content-Type' => 'application/json' ),
			'body'    => wp_json_encode( $body ),
		)
	);
	if ( is_wp_error( $res ) ) {
		return $res;
	}
	$code = wp_remote_retrieve_response_code( $res );
	if ( 429 === $code ) {
		return new WP_Error( 'findable_rate', __( 'Rate limited (10 scans per minute). Wait a moment and retry.', 'findable' ) );
	}
	$data = json_decode( wp_remote_retrieve_body( $res ), true );
	if ( 200 !== $code || ! is_array( $data ) || ! isset( $data['seo'] ) ) {
		return new WP_Error( 'findable_api', __( 'The findable API returned an unexpected response.', 'findable' ) );
	}

	$failing = array();
	foreach ( array( 'seo', 'aeo', 'geo' ) as $group ) {
		if ( empty( $data[ $group ]['checks'] ) || ! is_array( $data[ $group ]['checks'] ) ) {
			continue;
		}
		foreach ( $data[ $group ]['checks'] as $check ) {
			if ( isset( $check['pass'] ) && ! $check['pass'] && ( ! isset( $check['applicable'] ) || $check['applicable'] ) ) {
				if ( count( $failing ) >= 100 ) {
					break 2;
				}
				$failing[] = array(
					'group'    => $group,
					'label'    => sanitize_text_field( isset( $check['label'] ) ? $check['label'] : '' ),
					'severity' => sanitize_key( isset( $check['severity'] ) ? $check['severity'] : '' ),
				);
			}
		}
	}
	$score   = function ( $g ) use ( $data ) {
		return isset( $data[ $g ]['score'] ) ? (int) $data[ $g ]['score'] : null;
	};
	$summary = array(
		'overall'    => isset( $data['overall'] ) ? (int) $data['overall'] : null,
		'seo'        => $score( 'seo' ),
		'aeo'        => $score( 'aeo' ),
		'geo'        => $score( 'geo' ),
		'failing'    => $failing,
		'scanned_at' => time(),
		'url'        => $url,
	);
	update_post_meta( $post_id, '_findable_scan', $summary );
	return $summary;
}

function findable_report_url( $url ) {
	$parts = wp_parse_url( $url );
	$path  = isset( $parts['path'] ) ? rtrim( $parts['path'], '/' ) : '';
	return 'https://findable.page/' . ( isset( $parts['host'] ) ? $parts['host'] : '' ) . $path;
}

function findable_render_summary( $s ) {
	if ( ! is_array( $s ) ) {
		return '<p>' . esc_html__( 'Not scanned yet.', 'findable' ) . '</p>';
	}
	$fmt = function ( $v ) {
		return null === $v ? '-' : (int) $v;
	};
	$out  = '<p><strong>' . esc_html__( 'Overall', 'findable' ) . ': ' . esc_html( $fmt( $s['overall'] ) ) . '</strong><br>';
	$out .= 'SEO ' . esc_html( $fmt( $s['seo'] ) ) . ' &middot; AEO ' . esc_html( $fmt( $s['aeo'] ) ) . ' &middot; GEO ' . esc_html( $fmt( $s['geo'] ) ) . '</p>';
	if ( ! empty( $s['failing'] ) && is_array( $s['failing'] ) ) {
		$out .= '<p>' . esc_html__( 'Failing checks', 'findable' ) . ':</p><ul style="list-style:disc;margin-left:18px">';
		foreach ( $s['failing'] as $f ) {
			$out .= '<li>[' . esc_html( strtoupper( $f['group'] ) ) . '] ' . esc_html( $f['label'] ) . '</li>';
		}
		$out .= '</ul>';
	}
	$out .= '<p>' . esc_html(
		sprintf(
			/* translators: %s: date and time */
			__( 'Scanned %s', 'findable' ),
			wp_date( get_option( 'date_format' ) . ' ' . get_option( 'time_format' ), (int) $s['scanned_at'] )
		)
	) . ' &middot; <a href="' . esc_url( findable_report_url( $s['url'] ) ) . '" target="_blank" rel="noopener">' . esc_html__( 'Full report', 'findable' ) . '</a></p>';
	return $out;
}

add_action(
	'add_meta_boxes',
	function () {
		foreach ( array( 'post', 'page' ) as $type ) {
			add_meta_box( 'findable', 'findable', 'findable_meta_box', $type, 'side', 'default' );
		}
	}
);

function findable_meta_box( $post ) {
	wp_nonce_field( 'findable_save_' . $post->ID, 'findable_kp_nonce' );
	$kp = get_post_meta( $post->ID, '_findable_keyphrase', true );
	echo '<p><label for="findable-keyphrase">' . esc_html__( 'Keyphrase (optional)', 'findable' ) . '</label>';
	echo '<input type="text" class="widefat" id="findable-keyphrase" name="findable_keyphrase" value="' . esc_attr( $kp ) . '"></p>';
		echo '<p><button type="button" class="button" id="findable-scan" data-post="' . esc_attr( $post->ID ) . '">' . esc_html__( 'Scan with findable', 'findable' ) . '</button> <span class="spinner" id="findable-spinner" style="float:none"></span></p>';
	echo '<div id="findable-result">' . findable_render_summary( get_post_meta( $post->ID, '_findable_scan', true ) ) . '</div>'; // phpcs:ignore WordPress.Security.EscapeOutput -- escaped in renderer.
}

add_action(
	'save_post',
	function ( $post_id ) {
		if ( ! isset( $_POST['findable_kp_nonce'] ) || ! wp_verify_nonce( sanitize_text_field( wp_unslash( $_POST['findable_kp_nonce'] ) ), 'findable_save_' . $post_id ) ) {
			return;
		}
		if ( defined( 'DOING_AUTOSAVE' ) && DOING_AUTOSAVE ) {
			return;
		}
		if ( ! current_user_can( 'edit_post', $post_id ) ) {
			return;
		}
		$kp = isset( $_POST['findable_keyphrase'] ) ? sanitize_text_field( wp_unslash( $_POST['findable_keyphrase'] ) ) : '';
		update_post_meta( $post_id, '_findable_keyphrase', $kp );
	}
);

add_action(
	'wp_ajax_findable_scan',
	function () {
		check_ajax_referer( 'findable_scan', 'nonce' );
		$post_id = isset( $_POST['post_id'] ) ? absint( $_POST['post_id'] ) : 0;
		if ( ! $post_id || ! current_user_can( 'edit_post', $post_id ) ) {
			wp_send_json_error( array( 'message' => __( 'Not allowed.', 'findable' ) ), 403 );
		}
		if ( isset( $_POST['keyphrase'] ) ) {
			$kp = sanitize_text_field( wp_unslash( $_POST['keyphrase'] ) );
			update_post_meta( $post_id, '_findable_keyphrase', $kp );
		} else {
			$kp = (string) get_post_meta( $post_id, '_findable_keyphrase', true );
		}
		$lock = 'findable_lock_' . get_current_user_id();
		if ( get_transient( $lock ) ) {
			wp_send_json_error( array( 'message' => __( 'Scanning too fast. Wait a few seconds and retry.', 'findable' ) ), 429 );
		}
		set_transient( $lock, 1, 6 );
		$res = findable_scan_post( $post_id, $kp );
		if ( is_wp_error( $res ) ) {
			wp_send_json_error( array( 'message' => $res->get_error_message() ) );
		}
		wp_send_json_success(
			array(
				'html'    => findable_render_summary( $res ),
				'overall' => $res['overall'],
				'seo'     => $res['seo'],
				'aeo'     => $res['aeo'],
				'geo'     => $res['geo'],
				'failing' => count( $res['failing'] ),
				'report'  => findable_report_url( $res['url'] ),
			)
		);
	}
);

add_action(
	'admin_enqueue_scripts',
	function ( $hook ) {
		if ( ! in_array( $hook, array( 'post.php', 'post-new.php', 'tools_page_findable' ), true ) ) {
			return;
		}
		wp_enqueue_script( 'findable-admin', plugins_url( 'assets/admin.js', __FILE__ ), array(), FINDABLE_VERSION, true );
		wp_localize_script(
			'findable-admin',
			'findableData',
			array(
				'ajaxUrl' => admin_url( 'admin-ajax.php' ),
				'nonce'   => wp_create_nonce( 'findable_scan' ),
				'delayMs' => 7000,
				'i18n'    => array(
					'scanning' => __( 'Scanning...', 'findable' ),
					'failed'   => __( 'Scan failed.', 'findable' ),
					'report'   => __( 'Full report', 'findable' ),
					'done'     => __( 'Done.', 'findable' ),
				),
			)
		);
	}
);

foreach ( array( 'post', 'page' ) as $findable_type ) {
	add_filter(
		"manage_{$findable_type}s_columns",
		function ( $cols ) {
			$cols['findable'] = 'findable';
			return $cols;
		}
	);
	add_action(
		"manage_{$findable_type}s_custom_column",
		function ( $col, $post_id ) {
			if ( 'findable' !== $col ) {
				return;
			}
			$s = get_post_meta( $post_id, '_findable_scan', true );
			echo is_array( $s ) && null !== $s['overall'] ? esc_html( (int) $s['overall'] ) : '&ndash;';
		},
		10,
		2
	);
}

add_action(
	'admin_menu',
	function () {
		add_management_page( 'findable', 'findable', 'edit_published_posts', 'findable', 'findable_tools_page' );
	}
);

function findable_tools_page() {
	if ( ! current_user_can( 'edit_published_posts' ) ) {
		return;
	}
	$n = isset( $_GET['n'] ) ? min( 20, max( 1, absint( $_GET['n'] ) ) ) : 10; // phpcs:ignore WordPress.Security.NonceVerification -- read-only display value.
	$posts = get_posts(
		array(
			'post_type'      => 'post',
			'post_status'    => 'publish',
			'posts_per_page' => $n,
			'orderby'        => 'date',
			'order'          => 'DESC',
			'fields'         => 'ids',
		)
	);
	$posts = array_filter(
		$posts,
		function ( $id ) {
			return current_user_can( 'edit_post', $id );
		}
	);
	echo '<div class="wrap"><h1>findable</h1>';
	echo '<form method="get"><input type="hidden" name="page" value="findable">';
	echo '<label>' . esc_html__( 'Latest posts to scan (max 20)', 'findable' ) . ' <input type="number" name="n" min="1" max="20" value="' . esc_attr( $n ) . '"></label> ';
	echo '<button class="button">' . esc_html__( 'Set', 'findable' ) . '</button></form>';
	echo '<p><button type="button" class="button button-primary" id="findable-bulk">' . esc_html__( 'Scan latest posts', 'findable' ) . '</button> <span id="findable-bulk-status"></span></p>';
	echo '<p class="description">' . esc_html__( 'Scans run one at a time, 7 seconds apart, to respect the API rate limit.', 'findable' ) . '</p>';
	echo '<table class="widefat striped"><thead><tr><th>' . esc_html__( 'Post', 'findable' ) . '</th><th>' . esc_html__( 'Overall', 'findable' ) . '</th><th>SEO</th><th>AEO</th><th>GEO</th><th>' . esc_html__( 'Failing', 'findable' ) . '</th><th></th></tr></thead><tbody id="findable-bulk-body">';
	foreach ( $posts as $id ) {
		$s = get_post_meta( $id, '_findable_scan', true );
		$g = function ( $k ) use ( $s ) {
			return is_array( $s ) && null !== $s[ $k ] ? esc_html( (int) $s[ $k ] ) : '&ndash;';
		};
		echo '<tr data-post="' . esc_attr( $id ) . '"><td>' . esc_html( get_the_title( $id ) ) . '</td>';
		echo '<td data-f="overall">' . $g( 'overall' ) . '</td><td data-f="seo">' . $g( 'seo' ) . '</td><td data-f="aeo">' . $g( 'aeo' ) . '</td><td data-f="geo">' . $g( 'geo' ) . '</td>'; // phpcs:ignore WordPress.Security.EscapeOutput -- escaped in closure.
		$failing_count = is_array( $s['failing'] ?? null ) ? count( $s['failing'] ) : 0;
		echo '<td data-f="failing">' . ( is_array( $s ) ? esc_html( $failing_count ) : '&ndash;' ) . '</td><td data-f="report">';
		if ( is_array( $s ) ) {
			echo '<a href="' . esc_url( findable_report_url( $s['url'] ) ) . '" target="_blank" rel="noopener">' . esc_html__( 'Full report', 'findable' ) . '</a>';
		}
		echo '</td></tr>';
	}
	echo '</tbody></table></div>';
}
